package outputs

import (
	"fmt"
	"time"

	"github.com/aelsabbahy/goss/resource"
)

type Nagios struct{}

func (r Nagios) Output(results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	testCount := 0
	failed := 0
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if !testResult.Successful {
				failed++
			}
			testCount++
		}
	}

	duration := time.Now().Sub(startTime)
	if failed > 0 {
		fmt.Printf("GOSS CRITICAL - Count: %d, Failed: %d, Duration: %s\n", testCount, failed, duration)
		return 2
	}
	fmt.Printf("GOSS OK - Count: %d, Failed: %d, Duration: %s\n", testCount, failed, duration)
	return 0
}

func init() {
	RegisterOutputer("nagios", &Nagios{})
}
