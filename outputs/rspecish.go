package outputs

import (
	"fmt"

	"github.com/aelsabbahy/goss/resource"
	"github.com/fatih/color"
)

type Rspecish struct{}

func (r Rspecish) Output(results <-chan []resource.TestResult) (hasFail bool) {
	testCount := 0
	var failed []resource.TestResult
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if testResult.Result {
				fmt.Printf(green("."))
				testCount++
			} else {
				fmt.Printf(red("F"))
				failed = append(failed, testResult)
				testCount++
			}
		}
	}

	if len(failed) > 0 {
		color.Red("\n\nFailures:")
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
	RegisterOutputer("rspecish", &Rspecish{})
}
