package goss

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

// RuntimeConfig is configuration for various aspects of the goss system and
// is modeled on the cli context used in the cli tooling.  Mainly the translation
// into this format is there to support making the system usable as a package
type RuntimeConfig struct {
	FormatOptions     []string
	Vars              string
	VarsInline        string
	Spec              string
	Sleep             time.Duration
	RetryTimeout      time.Duration
	Cache             time.Duration
	Timeout           time.Duration
	MaxConcurrent     int
	NoColor           *bool
	OutputFormat      string
	PackageManager    string
	Endpoint          string
	ListenAddress     string
	ExcludeAttributes []string
	Insecure          bool
	NoFollowRedirects bool
	Server            string
	Username          string
	Password          string
	Debug             bool
}

func getGossConfig(vars string, varsInline string, specFile string) (cfg *GossConfig, err error) {
	// handle stdin
	var fh *os.File
	var path, source string
	var gossConfig GossConfig
	currentTemplateFilter = NewTemplateFilter(vars, varsInline)

	if specFile == "-" {
		source = "STDIN"
		fh = os.Stdin
		data, err := ioutil.ReadAll(fh)
		if err != nil {
			return nil, err
		}
		outStoreFormat = getStoreFormatFromData(data)
		gossConfig = ReadJSONData(data, true)
	} else {
		source = specFile
		path = filepath.Dir(specFile)
		outStoreFormat = getStoreFormatFromFileName(specFile)
		gossConfig = ReadJSON(specFile)
	}

	gossConfig = mergeJSONData(gossConfig, 0, path)

	if len(gossConfig.Resources()) == 0 {
		return nil, fmt.Errorf("found 0 tests, source: %v", source)
	}

	return &gossConfig, nil
}

func getOutputer(c *bool, format string) outputs.Outputer {
	if c != nil && *c {
		color.NoColor = true
	}
	if c != nil && !*c {
		color.NoColor = false
	}
	return outputs.GetOutputer(format)
}

func Validate(c *RuntimeConfig, startTime time.Time) (code int, err error) {
	outputConfig := util.OutputConfig{
		FormatOptions: c.FormatOptions,
	}

	gossConfig, err := getGossConfig(c.Vars, c.VarsInline, c.Spec)
	if err != nil {
		return 1, err
	}

	sys := system.New(c.PackageManager)
	outputer := getOutputer(c.NoColor, c.OutputFormat)

	sleep := c.Sleep
	retryTimeout := c.RetryTimeout
	i := 1
	for {
		iStartTime := time.Now()
		out := validate(sys, *gossConfig, c.MaxConcurrent)
		exitCode := outputer.Output(os.Stdout, out, iStartTime, outputConfig)
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

func validate(sys *system.System, gossConfig GossConfig, maxConcurrent int) <-chan []resource.TestResult {
	out := make(chan []resource.TestResult)
	in := make(chan resource.Resource)

	go func() {
		for _, t := range gossConfig.Resources() {
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
