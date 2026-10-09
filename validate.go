package goss

import (
	"bytes"
	"context"
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

func getGossConfig(varsFiles []string, varsInline string, specFile string) (cfg *GossConfig, err error) {
	// handle stdin
	var fh *os.File
	var path, source string
	var gossConfig GossConfig

	tf, err := NewTemplateFilter(varsFiles, varsInline)
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
		format, err := getStoreFormatFromData(data)
		if err != nil {
			return nil, err
		}
		setStoreFormat(format)

		gossConfig, err = ReadJSONData(data, true)
		if err != nil {
			return nil, err
		}
	} else {
		source = specFile
		path = filepath.Dir(specFile)
		format, err := getStoreFormatFromFileName(specFile)
		if err != nil {
			return nil, err
		}
		setStoreFormat(format)

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
		outputs.SetNoColor(true)
	}
	if c != nil && !*c {
		outputs.SetNoColor(false)
	}

	return outputs.GetOutputer(format)
}

// ValidateResults performs validation and provides programmatic access to validation results
// no retries or outputs are supported
func ValidateResults(ctx context.Context, c *util.Config) (results <-chan []resource.TestResult, err error) {
	gossConfig, err := getGossConfig(c.VarsFiles, c.VarsInline, c.Spec)
	if err != nil {
		return nil, err
	}

	sys := system.New(c.PackageManager)

	return validate(ctx, sys, *gossConfig, c.DisabledResourceTypes, c.MaxConcurrent), nil
}

// Validate performs validation, writes formatted output to stdout by default
// and supports retries and more, this is the full featured Validate used
// by the typical CLI invocation and will produce output to StdOut.  Use
// ValidateResults for programmatic access
func Validate(ctx context.Context, c *util.Config) (code int, err error) {
	err = setLogLevel(c)
	if err != nil {
		return 1, err
	}
	gossConfig, err := getGossConfig(c.VarsFiles, c.VarsInline, c.Spec)
	if err != nil {
		return 78, err
	}
	return ValidateConfig(ctx, c, gossConfig)
}

func ValidateConfig(ctx context.Context, c *util.Config, gossConfig *GossConfig) (code int, err error) {
	// Needed for contains-elements
	// Maybe we don't use this and use custom
	// contain_element_matcher is needed because it's single entry to avoid
	// transform message
	format.UseStringerRepresentation = true
	outputConfig := util.OutputConfig{
		FormatOptions: c.FormatOptions,
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

	return runValidationAttempts(
		c,
		outputer,
		ofh,
		outputConfig,
		func() <-chan []resource.TestResult {
			return validate(ctx, sys, *gossConfig, c.DisabledResourceTypes, c.MaxConcurrent)
		},
		func() {
			sys = system.New(c.PackageManager)
		},
		time.Now,
		time.Sleep,
		os.Stderr,
	)
}

func runValidationAttempts(
	c *util.Config,
	outputer outputs.Outputer,
	ofh io.Writer,
	outputConfig util.OutputConfig,
	validateAttempt func() <-chan []resource.TestResult,
	resetSystem func(),
	now func() time.Time,
	sleepFor func(time.Duration),
	progress io.Writer,
) (int, error) {
	sleep := c.Sleep
	retryTimeout := c.RetryTimeout
	discoveryRetry := c.OutputFormat == "discovery" && retryTimeout > 0
	attemptNumber := 1
	startTime := now()
	for {
		attemptWriter := ofh
		var attempt bytes.Buffer
		if discoveryRetry {
			attemptWriter = &attempt
		}

		exitCode := outputer.Output(attemptWriter, validateAttempt(), outputConfig)
		if discoveryRetry {
			switch exitCode {
			case 0:
				if err := writeValidationOutput(ofh, attempt.Bytes()); err != nil {
					fmt.Fprintf(progress, "discovery format: output failure: %v\n", err)
					return 2, nil
				}
				return 0, nil
			case 1:
				// A validation result may change on the next attempt.
			default:
				// Contract and encoding failures cannot be repaired by retrying.
				return exitCode, nil
			}
		}
		if retryTimeout == 0 || exitCode == 0 {
			return exitCode, nil
		}
		elapsed := now().Sub(startTime)
		if elapsed+sleep > retryTimeout {
			if discoveryRetry {
				if err := writeValidationOutput(ofh, attempt.Bytes()); err != nil {
					fmt.Fprintf(progress, "discovery format: output failure: %v\n", err)
					return 2, nil
				}
				fmt.Fprintf(progress, "discovery format: timeout of %s reached before tests entered a passing state\n", retryTimeout)
				return exitCode, nil
			}
			return 3, fmt.Errorf("timeout of %s reached before tests entered a passing state", retryTimeout)
		}
		if discoveryRetry {
			fmt.Fprintf(progress, "Retrying in %s (elapsed/timeout time: %.3fs/%s)\n", sleep, elapsed.Seconds(), retryTimeout)
		} else {
			color.Red("Retrying in %s (elapsed/timeout time: %.3fs/%s)\n\n\n", sleep, elapsed.Seconds(), retryTimeout)
		}
		// Reset cache
		resetSystem()
		sleepFor(sleep)
		attemptNumber++
		if discoveryRetry {
			fmt.Fprintf(progress, "Attempt #%d:\n", attemptNumber)
		} else {
			fmt.Printf("Attempt #%d:\n", attemptNumber)
		}
	}
}

func writeValidationOutput(w io.Writer, document []byte) error {
	written, err := w.Write(document)
	if err != nil {
		return err
	}
	if written != len(document) {
		return io.ErrShortWrite
	}
	return nil
}

func validate(ctx context.Context, sys *system.System, gossConfig GossConfig, skipList []string, maxConcurrent int) <-chan []resource.TestResult {
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
				out <- f.Validate(ctx, sys)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
