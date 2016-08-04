package goss

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

func getGossConfig(c *cli.Context) GossConfig {
	// handle stdin
	var fh *os.File
	var err error
	var path, source string
	specFile := c.GlobalString("gossfile")
	if specFile == "-" {
		source = "STDIN"
		fh = os.Stdin
	} else {
		source = specFile
		path = filepath.Dir(specFile)
		fh, err = os.Open(specFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
	data, err := ioutil.ReadAll(fh)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	gossConfig := mergeJSONData(ReadJSONData(data), 0, path)
	if len(gossConfig.Resources()) == 0 {
		fmt.Printf("Error: found 0 tests, source: %v\n", source)
		os.Exit(1)
	}
	return gossConfig
}

func getOutputer(c *cli.Context) outputs.Outputer {
	if c.Bool("no-color") {
		color.NoColor = true
	}
	return outputs.GetOutputer(c.String("format"))
}

func Validate(c *cli.Context, startTime time.Time) {
	gossConfig := getGossConfig(c)
	sys := system.New(c)
	outputer := getOutputer(c)

	sleep := c.Duration("sleep")
	retryTimeout := c.Duration("retry-timeout")
	i := 1
	for {
		iStartTime := time.Now()
		out := validate(sys, gossConfig)
		exitCode := outputer.Output(os.Stdout, out, iStartTime)
		if retryTimeout == 0 || exitCode == 0 {
			os.Exit(exitCode)
		}
		elapsed := time.Since(startTime)
		if elapsed+sleep > retryTimeout {
			color.Red("\nERROR: Timeout of %s reached before tests entered a passing state", retryTimeout)
			os.Exit(3)
		}
		color.Red("Retrying in %s (elapsed/timeout time: %.3fs/%s)\n\n\n", sleep, elapsed.Seconds(), retryTimeout)
		// Reset cache
		sys = system.New(c)
		time.Sleep(sleep)
		i++
		fmt.Printf("Attempt #%d:\n", i)
	}
}

func validate(sys *system.System, gossConfig GossConfig) <-chan []resource.TestResult {
	out := make(chan []resource.TestResult)
	in := make(chan resource.Resource)

	go func() {
		for _, t := range gossConfig.Resources() {
			in <- t
		}
		close(in)
	}()

	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	workerCount := runtime.NumCPU() * 5
	if workerCount > 50 {
		workerCount = 50
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
