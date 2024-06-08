package outputs

import (
	"fmt"
	"io"
	"log"
	"strings"
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
				logTrace("TRACE", "SUCCESS", testResult, false)
				fmt.Fprint(w, green("."))
			case resource.SKIP:
				logTrace("TRACE", "SKIP", testResult, false)
				fmt.Fprint(w, yellow("S"))
				failedOrSkippedGroup = append(failedOrSkippedGroup, testResult)
				skipped++
			case resource.FAIL:
				logTrace("TRACE", "FAIL", testResult, false)
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
	resstr := strings.ReplaceAll(outstr, "\n", " ")
	if failed > 0 {
		log.Printf("[DEBUG] FAIL SUMMARY: %s", resstr)
		return 1
	}
	log.Printf("[DEBUG] OK SUMMARY: %s", resstr)
	return 0
}
