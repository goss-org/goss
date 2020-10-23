package outputs

import (
	"fmt"
	"io"
	"log"
	"strings"
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
				logTrace("TRACE", "SUCCESS", testResult, false)
				fmt.Fprintf(w, green("."))
			case resource.SKIP:
				logTrace("TRACE", "SKIP", testResult, false)
				fmt.Fprintf(w, yellow("S"))
				failedOrSkippedGroup = append(failedOrSkippedGroup, testResult)
				skipped++
			case resource.FAIL:
				logTrace("WARN", "FAIL", testResult, false)
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
	fmt.Fprint(w, failedOrSkippedSummary(failedOrSkipped))

	outstr := summary(startTime, testCount, failed, skipped)
	fmt.Fprint(w, outstr)
	resstr := strings.ReplaceAll(outstr, "\n", " ")
	if failed > 0 {
		log.Printf("[WARN] FAIL SUMMARY: %s", resstr)
		return 1
	}
	log.Printf("[INFO] OK SUMMARY: %s", resstr)
	return 0
}
