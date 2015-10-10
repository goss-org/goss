package outputs

import (
	"fmt"

	"github.com/aelsabbahy/goss/resource"
)

type Outputer interface {
	Output(<-chan resource.TestResult) bool
}

type Rspecish struct{}

func (r Rspecish) Output(results <-chan resource.TestResult) (hasFail bool) {
	testCount := 0
	var failed []resource.TestResult
	for testResult := range results {
		//fmt.Printf("%v: %s.\n", testResult.Duration, testResult.Desc)
		if testResult.Result {
			fmt.Printf(".")
			testCount++
		} else {
			fmt.Printf("F")
			failed = append(failed, testResult)
			testCount++
		}
	}

	for _, testResult := range failed {
		fmt.Printf("\n%s\n", testResult.Desc)
	}

	fmt.Printf("\n\nCount: %d failed: %d\n", testCount, len(failed))
	if len(failed) > 0 {
		return true
	}
	return false
}
