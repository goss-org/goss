package outputs

import (
	"fmt"
	"io"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
)

type Documentation struct{}

func (r Documentation) ValidOptions() []*formatOption {
	return []*formatOption{}
}

func (r Documentation) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {
	includeRaw := util.IsValueInList(foIncludeRaw, outConfig.FormatOptions)

	testCount := 0
	var failedOrSkipped [][]resource.TestResult
	var skipped, failed int
	for resultGroup := range results {
		failedOrSkippedGroup := []resource.TestResult{}
		first := resultGroup[0]
		header := header(first)
		if header != "" {
			fmt.Fprint(w, header)
		}
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.SUCCESS:
				fmt.Fprintln(w, humanizeResult(testResult, false, includeRaw))
			case resource.SKIP:
				fmt.Fprintln(w, humanizeResult(testResult, false, includeRaw))
				failedOrSkippedGroup = append(failedOrSkippedGroup, testResult)
				skipped++
			case resource.FAIL:
				fmt.Fprintln(w, humanizeResult(testResult, false, includeRaw))
				failedOrSkippedGroup = append(failedOrSkippedGroup, testResult)
				failed++
			}
			testCount++
		}
		if len(failedOrSkippedGroup) > 0 {
			failedOrSkipped = append(failedOrSkipped, failedOrSkippedGroup)
		}
	}

	fmt.Fprint(w, "\n\n")
	fmt.Fprint(w, failedOrSkippedSummary(failedOrSkipped, includeRaw))

	fmt.Fprint(w, summary(startTime, testCount, failed, skipped))
	if failed > 0 {
		return 1
	}
	return 0
}
