package outputs

import (
	"github.com/aelsabbahy/goss/resource"
	"github.com/fatih/color"
)

type Documentation struct {
	color bool
}

func (d *Documentation) SetColor(t bool) {
	d.color = t
}

func (r Documentation) Output(results <-chan resource.TestResult) (hasFail bool) {
	testCount := 0
	var failed []resource.TestResult
	//var lastSeen string
	for testResult := range results {
		// Not sure if I want this or not
		//seenKey := fmt.Sprintf("%s-%s", testResult.ResourceType, testResult.Title)
		//if lastSeen != seenKey {
		//	fmt.Println("")
		//}
		//lastSeen = seenKey

		//fmt.Printf("%v: %s.\n", testResult.Duration, testResult.Desc)
		if testResult.Result {
			color.Green(humanizeResult(testResult))
			testCount++
		} else {
			color.Red(humanizeResult(testResult))
			failed = append(failed, testResult)
			testCount++
		}
	}

	if len(failed) > 0 {
		color.Red("\n\nFailures:")
		for _, testResult := range failed {
			color.Red(humanizeResult(testResult))
			//fmt.Printf("\n%s\n", testResult.Desc)
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
