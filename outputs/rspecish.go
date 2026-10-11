package outputs

import (
	"fmt"
	"io"
	"time"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

type Rspecish struct{}

func (r Rspecish) ValidOptions() []*formatOption {
	return []*formatOption{}
}

func (r Rspecish) Output(w io.Writer, results <-chan []resource.TestResult,
	outConfig util.OutputConfig) (exitCode int) {

	logger := util.LoggerOrDiscard(outConfig.Logger)

	sort := util.IsValueInList(foSort, outConfig.FormatOptions)
	results = getResults(results, sort)

	var startTime time.Time
	var endTime time.Time
	testCount := 0
	var failedOrSkipped [][]resource.TestResult
	var skipped, failed int
	for resultGroup := range results {
		failedOrSkippedGroup := []resource.TestResult{}
		for _, testResult := range resultGroup {
			// Calculates the start and end times based on the start of the first test
			// and end of the last test, this allows the time/duration to be stable
			// FIXME: move this to shared code
			if startTime.IsZero() || testResult.StartTime.Before(startTime) {
				startTime = testResult.StartTime
			}
			if endTime.IsZero() || testResult.EndTime.After(endTime) {
				endTime = testResult.EndTime
			}
			switch testResult.Result {
			case resource.SUCCESS:
				logTrace(logger, outcomeSuccess, testResult, false)
				fmt.Fprint(w, green("."))
			case resource.SKIP:
				logTrace(logger, outcomeSkip, testResult, false)
				fmt.Fprint(w, yellow("S"))
				failedOrSkippedGroup = append(failedOrSkippedGroup, testResult)
				skipped++
			case resource.FAIL:
				logTrace(logger, outcomeFail, testResult, false)
				fmt.Fprint(w, red("F"))
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
	includeRaw := !util.IsValueInList(foExcludeRaw, outConfig.FormatOptions)

	fmt.Fprint(w, failedOrSkippedSummary(failedOrSkipped, includeRaw))

	outstr := summary(startTime, endTime, testCount, failed, skipped)
	fmt.Fprint(w, outstr)

	// The counts and the duration were already the whole content of this
	// outputer's summary record; as attributes they arrive without the colour
	// escapes the printed version carries.
	status := statusOK
	exitCode = 0
	if failed > 0 {
		status = statusFail
		exitCode = 1
	}
	logger.Debug("validation summary",
		"status", status,
		"total", testCount,
		"failed", failed,
		"skipped", skipped,
		"duration_seconds", endTime.Sub(startTime).Seconds())

	return exitCode
}
