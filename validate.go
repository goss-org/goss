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

	"github.com/goss-org/goss/outputs"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

func getGossConfig(vars string, varsInline string, specFile string) (cfg *GossConfig, err error) {
	// handle stdin
	var fh *os.File
	var path, source string
	var gossConfig GossConfig

	currentTemplateFilter, err = NewTemplateFilter(vars, varsInline)
	if err != nil {
		return nil, err
	}

	if specFile == "-" {
		source = "STDIN"
		fh = os.Stdin
		data, err := io.ReadAll(fh)
		if err != nil {
			return nil, err
		}
		outStoreFormat, err = getStoreFormatFromData(data)
		if err != nil {
			return nil, err
		}

		gossConfig, err = ReadJSONData(data, true)
		if err != nil {
			return nil, err
		}
	} else {
		source = specFile
		path = filepath.Dir(specFile)
		outStoreFormat, err = getStoreFormatFromFileName(specFile)
		if err != nil {
			return nil, err
		}

		gossConfig, err = ReadJSON(specFile)
		if err != nil {
			return nil, err
		}
	}

	gossConfig, err = mergeJSONData(gossConfig, 0, path)
	if err != nil {
		return nil, err
	}

	if len(gossConfig.Resources()) == 0 {
		return nil, fmt.Errorf("found 0 tests, source: %v", source)
	}

	return &gossConfig, nil
}

func getOutputer(c *bool, format string) (outputs.Outputer, error) {
	if c != nil && *c {
		color.NoColor = true
	}
	if c != nil && !*c {
		color.NoColor = false
	}

	return outputs.GetOutputer(format)
}

// ValidateResults performs validation and provides programmatic access to validation results
// no retries or outputs are supported
func ValidateResults(c *util.Config) (results <-chan []resource.TestResult, err error) {
	gossConfig, err := getGossConfig(c.Vars, c.VarsInline, c.Spec)
	if err != nil {
		return nil, err
	}

	sys := system.New(c.PackageManager)

	return validate(sys, *gossConfig, c.DisabledResourceTypes, c.MaxConcurrent), nil
}

// Validate performs validation, writes formatted output to stdout by default
// and supports retries and more, this is the full featured Validate used
// by the typical CLI invocation and will produce output to StdOut.  Use
// ValidateResults for programmatic access
func Validate(c *util.Config, startTime time.Time) (code int, err error) {
	err = setLogLevel(c)
	if err != nil {
		return 1, err
	}
	outputConfig := util.OutputConfig{
		FormatOptions: c.FormatOptions,
	}

	gossConfig, err := getGossConfig(c.Vars, c.VarsInline, c.Spec)
	if err != nil {
		return 1, err
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
	for {
		iStartTime := time.Now()
		out := validate(sys, *gossConfig, c.DisabledResourceTypes, c.MaxConcurrent)
		exitCode := outputer.Output(ofh, out, iStartTime, outputConfig)
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

func validate(sys *system.System, gossConfig GossConfig, skipList []string, maxConcurrent int) <-chan []resource.TestResult {
	out := make(chan []resource.TestResult)
	in := make(chan resource.Resource)

	go func() {
		for _, t := range gossConfig.Resources() {
			if util.IsValueInList(t.TypeName(), skipList) || util.IsValueInList(t.TypeKey(), skipList) {
				t.SetSkip()
			}

			in <- t
		}
		close(in)
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
