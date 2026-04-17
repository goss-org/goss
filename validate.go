package goss

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/onsi/gomega/format"

	"github.com/goss-org/goss/outputs"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

// getGossConfig loads and merges gossfiles, emitting any merge warnings via
// c.Log(). It is an edge-layer function: it owns the translation between
// pure config-merging results (warnings returned as values) and the
// process's log sink.
func getGossConfig(c *util.Config, vars string, varsInline string, specFile string) (cfg *GossConfig, err error) {
	// handle stdin
	var fh *os.File
	var path, source string
	var gossConfig GossConfig

	tf, err := NewTemplateFilter(vars, varsInline)
	if err != nil {
		return nil, err
	}
	setTemplateFilter(tf)

	if specFile == "-" {
		source = "STDIN"
		fh = os.Stdin
		data, err := io.ReadAll(fh)
		if err != nil {
			return nil, err
		}
		storeFormat, err := getStoreFormatFromData(data)
		if err != nil {
			return nil, err
		}
		setStoreFormat(storeFormat)

		gossConfig, err = ReadJSONData(data, true)
		if err != nil {
			return nil, err
		}
	} else {
		source = specFile
		path = filepath.Dir(specFile)
		storeFormat, err := getStoreFormatFromFileName(specFile)
		if err != nil {
			return nil, err
		}
		setStoreFormat(storeFormat)

		gossConfig, err = ReadJSON(specFile)
		if err != nil {
			return nil, err
		}
	}

	gossConfig, warnings, err := mergeJSONData(gossConfig, 0, path)
	if err != nil {
		return nil, err
	}
	logWarnings(c.Log(), warnings)

	if len(gossConfig.Resources()) == 0 {
		return nil, fmt.Errorf("found 0 tests, source: %v", source)
	}

	return &gossConfig, nil
}

func getOutputer(c *bool, format string) (outputs.Outputer, error) {
	// color.NoColor is a package-level global that was historically set
	// directly here. To avoid races under parallel/serve workloads, we
	// initialize it at most once per process. If the caller explicitly
	// requested a value, honour it; otherwise leave the library's default
	// (derived from terminal detection) in place.
	if c != nil {
		util.InitNoColor(*c)
	}

	return outputs.GetOutputer(format)
}

// ValidateResults performs validation and provides programmatic access to validation results
// no retries or outputs are supported
func ValidateResults(c *util.Config) (results <-chan []resource.TestResult, err error) {
	gossConfig, err := getGossConfig(c, c.Vars, c.VarsInline, c.Spec)
	if err != nil {
		return nil, err
	}

	sys := system.New(c.PackageManager)

	return validate(sys, *gossConfig, c.DisabledResourceTypes, c.IncludeMarks, c.ExcludeMarks, c.MaxConcurrent, c.Log()), nil
}

// Validate performs validation, writes formatted output to stdout by default
// and supports retries and more, this is the full featured Validate used
// by the typical CLI invocation and will produce output to StdOut.  Use
// ValidateResults for programmatic access
func Validate(c *util.Config) (code int, err error) {
	err = setLogLevel(c)
	if err != nil {
		return 1, err
	}
	gossConfig, err := getGossConfig(c, c.Vars, c.VarsInline, c.Spec)
	if err != nil {
		return 78, err
	}
	return ValidateConfig(c, gossConfig)
}

// gomegaFormatOnce guards the single mutation of the gomega format package's
// UseStringerRepresentation flag. gomega/format stores this as a package
// global, and ValidateConfig may be invoked concurrently under `goss serve`,
// so we set it exactly once per process.
var gomegaFormatOnce sync.Once

func ValidateConfig(c *util.Config, gossConfig *GossConfig) (code int, err error) {
	// Needed for contains-elements
	// Maybe we don't use this and use custom
	// contain_element_matcher is needed because it's single entry to avoid
	// transform message.
	gomegaFormatOnce.Do(func() {
		format.UseStringerRepresentation = true
	})
	outputConfig := util.OutputConfig{
		FormatOptions: c.FormatOptions,
		Logger:        c.Logger,
	}

	sys := system.New(c.PackageManager)
	outputer, err := getOutputer(c.NoColor, c.OutputFormat)
	if err != nil {
		return 1, err
	}

	var ofh io.Writer
	ofh = os.Stdout
	if c.OutputWriter != nil {
		ofh = c.OutputWriter
	}

	sleep := c.Sleep
	retryTimeout := c.RetryTimeout
	i := 1
	startTime := time.Now()
	for {
		out := validate(sys, *gossConfig, c.DisabledResourceTypes, c.IncludeMarks, c.ExcludeMarks, c.MaxConcurrent, c.Log())
		exitCode := outputer.Output(ofh, out, outputConfig)
		if retryTimeout == 0 || exitCode == 0 {
			return exitCode, nil
		}
		elapsed := time.Since(startTime)
		if elapsed+sleep > retryTimeout {
			return 3, fmt.Errorf("timeout of %s reached before tests entered a passing state", retryTimeout)
		}
		color.Red("Retrying in %s (elapsed/timeout time: %.3fs/%s)\n\n\n", sleep, elapsed.Seconds(), retryTimeout)
		// Reset cache
		sys = system.New(c.PackageManager)
		time.Sleep(sleep)
		i++
		fmt.Printf("Attempt #%d:\n", i)
	}
}

func validate(sys *system.System, gossConfig GossConfig, skipList []string, includeMarks []string, excludeMarks []string, maxConcurrent int, logger util.Logger) <-chan []resource.TestResult {
	out := make(chan []resource.TestResult)
	in := make(chan resource.Resource)

	// Emit a single informational summary when mark filters are active.
	// This is gated behind the caller's log level (via logutils) by using
	// the [DEBUG] prefix, so it is quiet at the default INFO level and
	// opt-in when someone runs with -L DEBUG.
	logMarkFilterSummary(logger, gossConfig, includeMarks, excludeMarks)

	go func() {
		filteredByMarks := 0
		for _, t := range gossConfig.Resources() {
			if util.IsValueInList(t.TypeName(), skipList) || util.IsValueInList(t.TypeKey(), skipList) {
				t.SetSkip()
			}

			if !shouldRunByMarks(t.GetMarks(), includeMarks, excludeMarks) {
				t.SetSkip()
				filteredByMarks++
			}

			in <- t
		}
		close(in)

		if logger != nil && (len(includeMarks) > 0 || len(excludeMarks) > 0) {
			logger.Printf("[DEBUG] marks filter skipped %d of %d resources", filteredByMarks, len(gossConfig.Resources()))
		}
	}()

	workerCount := runtime.NumCPU() * 5
	if workerCount > maxConcurrent {
		workerCount = maxConcurrent
	}
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range in {
				out <- f.Validate(sys)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// shouldRunByMarks decides whether a resource with the given marks should run
// under the supplied include/exclude filters.
//
// Rules:
//   - If includeMarks is non-empty, the resource must have at least one matching
//     mark to run. Resources with no marks are excluded when includeMarks is set.
//   - If excludeMarks is non-empty, the resource is skipped if it has any
//     matching mark.
//   - When both are set, inclusion is evaluated first, then exclusion.
//   - When both are empty, the resource runs by default.
func shouldRunByMarks(resourceMarks, includeMarks, excludeMarks []string) bool {
	if len(includeMarks) > 0 {
		if !hasAnyMark(resourceMarks, includeMarks) {
			return false
		}
	}
	if len(excludeMarks) > 0 {
		if hasAnyMark(resourceMarks, excludeMarks) {
			return false
		}
	}
	return true
}

// logMarkFilterSummary emits a DEBUG line describing the active mark filters.
// It is a no-op when no filters are set or when logger is nil. The message is
// prefixed with [DEBUG] so hashicorp/logutils gates it by log level.
func logMarkFilterSummary(logger util.Logger, gossConfig GossConfig, includeMarks, excludeMarks []string) {
	if logger == nil {
		return
	}
	if len(includeMarks) == 0 && len(excludeMarks) == 0 {
		return
	}
	logger.Printf("[DEBUG] mark filters active: include=%v exclude=%v total_resources=%d",
		includeMarks, excludeMarks, len(gossConfig.Resources()))
}

// hasAnyMark returns true if any element of resourceMarks is also in filterMarks.
func hasAnyMark(resourceMarks, filterMarks []string) bool {
	if len(resourceMarks) == 0 || len(filterMarks) == 0 {
		return false
	}
	for _, rm := range resourceMarks {
		for _, fm := range filterMarks {
			if rm == fm {
				return true
			}
		}
	}
	return false
}
