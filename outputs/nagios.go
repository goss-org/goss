package outputs

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

type Nagios struct{}

func (r Nagios) ValidOptions() []*formatOption {
	return []*formatOption{
		{name: foPerfData},
		{name: foVerbose},
	}
}

func (r Nagios) Output(w io.Writer, results <-chan []resource.TestResult,
	outConfig util.OutputConfig) (exitCode int) {

	var testCount, failed, skipped int

	var perfdata, verbose bool
	perfdata = util.IsValueInList(foPerfData, outConfig.FormatOptions)
	verbose = util.IsValueInList(foVerbose, outConfig.FormatOptions)
	includeRaw := !util.IsValueInList(foExcludeRaw, outConfig.FormatOptions)
	specFile := filepath.Base(outConfig.SpecFile)

	var startTime time.Time
	var endTime time.Time
	var summary = make(map[int]string)

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if startTime.IsZero() || testResult.StartTime.Before(startTime) {
				startTime = testResult.StartTime
			}
			if endTime.IsZero() || testResult.EndTime.After(endTime) {
				endTime = testResult.EndTime
			}
			switch testResult.Result {
			case resource.FAIL:
				if util.IsValueInList(foVerbose, outConfig.FormatOptions) {
					summary[failed] = "Fail " + strconv.Itoa(failed+1) + " - " + humanizeResult(testResult, true, includeRaw) + "\n"
				}
				failed++
			case resource.SKIP:
				skipped++
			}
			testCount++
		}
	}

	duration := endTime.Sub(startTime)
	return writeNagiosResult(w, specFile, testCount, failed, skipped, duration, perfdata, verbose, summary)
}

func writeNagiosResult(w io.Writer, specFile string, testCount, failed, skipped int, duration time.Duration, perfdata, verbose bool, summary map[int]string) int {
	nagiosState := "OK"
	if failed > 0 {
		nagiosState = "CRITICAL"
	}

	fmt.Fprintf(w, "GOSS-%s %s - Count: %d, Failed: %d, Skipped: %d, Duration: %.3fs", specFile, nagiosState, testCount, failed, skipped, duration.Seconds())
	if perfdata {
		fmt.Fprintf(w, "|total=%d failed=%d skipped=%d duration=%.3fs", testCount, failed, skipped, duration.Seconds())
	}
	fmt.Fprint(w, "\n")
	if verbose {
		for i := 0; i < failed; i++ {
			fmt.Fprintf(w, "%s", summary[i])
		}
	}
	if failed > 0 {
		return 2
	} else {
		return 0
	}
}
