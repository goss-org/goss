package outputs

import (
	"fmt"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/fatih/color"
)

type Rspecish struct{}

func (r Rspecish) Output(results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	testCount := 0
	var failed []resource.TestResult
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if testResult.Successful {
				fmt.Printf(green("."))
			} else {
				fmt.Printf(red("F"))
				failed = append(failed, testResult)
			}
			testCount++
		}
	}

	fmt.Print("\n\n")
	if len(failed) > 0 {
		color.Red("Failures:")
		for _, testResult := range failed {
			fmt.Println(humanizeResult(testResult))
		}
		fmt.Print("\n")
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
	RegisterOutputer("rspecish", &Rspecish{})
}
