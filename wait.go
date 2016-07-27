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
	done := make(chan bool, 1)
	timedout := make(chan bool, 1)
	go func() {
		i := 0
		for {
			i++
			fmt.Printf("Attempt #%d:\n", i)
			exitCode := make(chan int, 1)
			go func() {
				exitCode <- validate(context.TODO(), sys, gossConfig, time.Now(), outputer)
			}()
			select {
			case e := <-exitCode:
				fmt.Println("wtf1")
				if e == 0 {
					done <- true
					return
				}
			case <-timedout:
				fmt.Println("wtf2")
				return
			}
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("\n\nRetrying in %s (elapsed time: %.3fs, timeout: %s)\n", sleep, elapsed, timeout)
			time.Sleep(sleep)
		}
	}()
	select {
	case <-time.After(timeout):
		timedout <- true
		//time.Sleep(10 * time.Millisecond)
		time.Sleep(500 * time.Millisecond)
		color.Red("\nERROR: Timeout of %s reached before tests entered a passing state", timeout)
		os.Exit(3)
	case <-done:
	}
	os.Exit(0)
}
