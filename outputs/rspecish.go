package outputs

import (
	"fmt"
	"time"

	"github.com/aelsabbahy/goss/resource"
)

type Rspecish struct{}

func (r Rspecish) Output(results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	testCount := 0
	var failedOrSkipped [][]resource.TestResult
	var skipped, failed int
	for resultGroup := range results {
		failedOrSkippedGroup := []resource.TestResult{}
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.SUCCESS:
				fmt.Printf(green("."))
			case resource.SKIP:
				fmt.Printf(yellow("S"))
				failedOrSkippedGroup = append(failedOrSkippedGroup, testResult)
				skipped++
			case resource.FAIL:
				fmt.Printf(red("F"))
				failedOrSkippedGroup = append(failedOrSkippedGroup, testResult)
				failed++
			}
			testCount++
		}
		if len(failedOrSkippedGroup) > 0 {
			failedOrSkipped = append(failedOrSkipped, failedOrSkippedGroup)
		}
	}

	fmt.Print("\n\n")
	fmt.Print(failedOrSkippedSummary(failedOrSkipped))

	fmt.Print(summary(startTime, testCount, failed, skipped))
	if failed > 0 {
		return 1
	}
	return 0
}

func init() {
	RegisterOutputer("rspecish", &Rspecish{})
}
