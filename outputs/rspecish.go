package outputs

import (
	"fmt"
	"io"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
)

type Rspecish struct{}

func (r Rspecish) ValidOptions() []*formatOption {
	return []*formatOption{}
}

func (r Rspecish) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {

	testCount := 0
	var failedOrSkipped [][]resource.TestResult
	var skipped, failed int
	for resultGroup := range results {
		failedOrSkippedGroup := []resource.TestResult{}
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.SUCCESS:
				fmt.Fprintf(w, green("."))
			case resource.SKIP:
				fmt.Fprintf(w, yellow("S"))
				failedOrSkippedGroup = append(failedOrSkippedGroup, testResult)
				skipped++
			case resource.FAIL:
				fmt.Fprintf(w, red("F"))
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
	includeRaw := util.IsValueInList("include_raw", outConfig.FormatOptions)

	fmt.Fprint(w, failedOrSkippedSummary(failedOrSkipped, includeRaw))

	fmt.Fprint(w, summary(startTime, testCount, failed, skipped))
	if failed > 0 {
		return 1
	}
	return 0
}
