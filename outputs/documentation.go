package outputs

import (
	"fmt"

	"github.com/aelsabbahy/goss/resource"
	"github.com/fatih/color"
)

type Documentation struct{}

func (r Documentation) Output(results <-chan []resource.TestResult) (hasFail bool) {
	testCount := 0
	var failed []resource.TestResult
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if testResult.Result {
				fmt.Println(humanizeResult(testResult))
				testCount++
			} else {
				fmt.Println(humanizeResult(testResult))
				failed = append(failed, testResult)
				testCount++
			}
		}
		fmt.Println("")
	}

	if len(failed) > 0 {
		color.Red("\nFailures:")
		for _, testResult := range failed {
			fmt.Println(humanizeResult(testResult))
		}
	}

	if len(failed) > 0 {
		color.Red("\n\nCount: %d failed: %d\n", testCount, len(failed))
		return true
	}
	color.Green("\n\nCount: %d failed: %d\n", testCount, len(failed))
	return false
}

func init() {
	RegisterOutputer("documentation", &Documentation{})
}
