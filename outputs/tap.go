package outputs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aelsabbahy/goss/resource"
)

type Tap struct{}

func (r Tap) Output(results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	testCount := 0
	failed := 0

	var summary map[int]string
	summary = make(map[int]string)

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.SUCCESS:
				summary[testCount] = "ok " + strconv.Itoa(testCount+1) + " - " + humanizeResult2(testResult) + "\n"
			case resource.FAIL:
				summary[testCount] = "not ok " + strconv.Itoa(testCount+1) + " - " + humanizeResult2(testResult) + "\n"
				failed++
			case resource.SKIP:
				summary[testCount] = "ok " + strconv.Itoa(testCount+1) + " - # SKIP " + humanizeResult2(testResult) + "\n"
			default:
				panic(fmt.Sprintf("Unexpected Result Code: %v\n", testResult.Result))
			}
			testCount++
		}
	}

	fmt.Printf("1..%d\n", testCount)

	for i := 0; i < testCount; i++ {
		fmt.Printf("%s", summary[i])
	}

	if failed > 0 {
		return 1
	}

	return 0
}

func init() {
	RegisterOutputer("tap", &Tap{})
}
