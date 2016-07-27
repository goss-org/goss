package goss

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"golang.org/x/net/context"

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
	var path string
	if !c.GlobalIsSet("gossfile") && hasStdin() {
		fh = os.Stdin
	} else {
		specFile := c.GlobalString("gossfile")
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

	exitCode := validate(context.TODO(), sys, gossConfig, startTime, outputer)
	os.Exit(exitCode)
}

func validate(ctx context.Context, sys *system.System, gossConfig GossConfig, startTime time.Time, outputer outputs.Outputer) int {
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

	exitCode := outputer.Output(out, startTime)
	return exitCode
}

func hasStdin() bool {
	if fi, err := os.Stdin.Stat(); err == nil {
		mode := fi.Mode()
		if (mode&os.ModeNamedPipe != 0) || mode.IsRegular() {
			return true
		}
	}
	return false
}
