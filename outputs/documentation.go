package outputs

import (
	"fmt"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/fatih/color"
)

type Documentation struct{}

func (r Documentation) Output(results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	testCount := 0
	var failed [][]resource.TestResult
	for resultGroup := range results {
		failedGroup := []resource.TestResult{}
		first := resultGroup[0]
		header := header(first)
		if header != "" {
			fmt.Print(header)
		}
		for _, testResult := range resultGroup {
			if testResult.Successful {
				fmt.Println(humanizeResult(testResult))
				testCount++
			} else {
				fmt.Println(humanizeResult(testResult))
				failedGroup = append(failedGroup, testResult)
				testCount++
			}
		}
		fmt.Println("")
		if len(failedGroup) > 0 {
			failed = append(failed, failedGroup)
		}
	}

	fmt.Print("\n\n")
	if len(failed) > 0 {
		fmt.Println("Failures:\n")
		for _, failedGroup := range failed {
			first := failedGroup[0]
			header := header(first)
			if header != "" {
				fmt.Print(header)
			}
			for _, testResult := range failedGroup {
				fmt.Println(humanizeResult(testResult))
			}
			fmt.Print("\n")
		}
	}

	fmt.Printf("Total Duration: %.3fs\n", time.Since(startTime).Seconds())
	if len(failed) > 0 {
		color.Red("Count: %d, Failed: %d\n", testCount, len(failed))
		return 1
	}
	color.Green("Count: %d, Failed: %d\n", testCount, len(failed))
	return 0
}

func init() {
	RegisterOutputer("documentation", &Documentation{})
}
