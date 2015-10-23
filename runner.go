package goss

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

func Run(specFile string, c *cli.Context) {
	sys := system.New(c)

	// handle stdin
	var fh *os.File
	var err error
	var path string
	if hasStdin() {
		fh = os.Stdin
	} else {
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
	configJSON := mergeJSONData(ReadJSONData(data), 0, path)

	out := make(chan []resource.TestResult)

	in := make(chan resource.Resource)

	go func() {
		for _, t := range configJSON.Resources() {
			in <- t
		}
		close(in)
	}()

	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	gomaxprocs := runtime.GOMAXPROCS(-1)
	workerCount := gomaxprocs * 5
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

	//var outputer outputs.Outputer
	if c.Bool("no-color") {
		color.NoColor = true
	}

	outputer := outputs.GetOutputer(c.String("format"))

	if hasFail := outputer.Output(out); hasFail {
		os.Exit(1)
	}

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
