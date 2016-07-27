package goss

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/aelsabbahy/goss/system"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

func Wait(c *cli.Context) {
	startTime := time.Now()
	gossConfig := getGossConfig(c)
	sys := system.New(c)
	sleep := c.Duration("sleep")
	timeout := c.Duration("timeout")
	outputer := getOutputer(c)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	i := 0
	for {
		i++
		fmt.Printf("Attempt #%d:\n", i)
		iStartTime := time.Now()
		out := validate(ctx, sys, gossConfig)
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				color.Red("\nERROR: Timeout of %s reached before tests entered a passing state", timeout)
				os.Exit(3)
			}
		default:
			exitCode := outputer.Output(out, iStartTime)
			if exitCode == 0 {
				os.Exit(0)
			}
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("\n\nRetrying in %s (elapsed time: %.3fs, timeout: %s)\n", sleep, elapsed, timeout)
			time.Sleep(sleep)
		}
	}
}
