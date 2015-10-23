package outputs

import (
	"github.com/aelsabbahy/goss/resource"
	"github.com/fatih/color"
)

type Rspecish struct {
	color bool
}

func (r *Rspecish) SetColor(t bool) {
	r.color = t
}

func (r Rspecish) Output(results <-chan []resource.TestResult) (hasFail bool) {
	green := color.New(color.FgGreen).PrintfFunc()
	red := color.New(color.FgRed).PrintfFunc()
	testCount := 0
	var failed []resource.TestResult
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if testResult.Result {
				green(".")
				testCount++
			} else {
				red("F")
				failed = append(failed, testResult)
				testCount++
			}
		}
	}

	if len(failed) > 0 {
		color.Red("\n\nFailures:")
		for _, testResult := range failed {
			color.Red(humanizeResult(testResult))
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
