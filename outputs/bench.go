package outputs

import (
	"fmt"
	"io"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
)

type Bench struct{}

func (r Bench) ValidOptions() []*formatOption {
	return []*formatOption{
		{name: foSort},
	}
}

func (r Bench) Output(w io.Writer, results <-chan []resource.TestResult, outConfig util.OutputConfig) (exitCode int) {
	includeRaw := util.IsValueInList(foIncludeRaw, outConfig.FormatOptions)

	sort := util.IsValueInList(foSort, outConfig.FormatOptions)
	results = getResults(results, sort)

	var startTime time.Time
	var endTime time.Time
	var testCount, skipped, failed int
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if startTime.IsZero() || testResult.StartTime.Before(startTime) {
				startTime = testResult.StartTime
			}
			if endTime.IsZero() || testResult.EndTime.After(endTime) {
				endTime = testResult.EndTime
			}
			fmt.Fprintf(w, "%v %s\n", testResult.Duration, humanizeResult(testResult, true, includeRaw))
			switch testResult.Result {
			case resource.SKIP:
				skipped++
			case resource.FAIL:
				failed++
			}
			testCount++
		}
	}

	fmt.Fprint(w, summary(startTime, endTime, testCount, failed, skipped))
	if failed > 0 {
		return 1
	}
	return 0
}
