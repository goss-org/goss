package outputs

import (
	"fmt"
	"io"
	"time"

	"github.com/aelsabbahy/goss/resource"
)

type Nagios struct{}

func (r Nagios) SetReportURL(stringified string) error { return nil }

func (r Nagios) Output(w io.Writer, results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	var testCount, failed, skipped int
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.FAIL:
				failed++
			case resource.SKIP:
				skipped++
			}
			testCount++
		}
	}

	duration := time.Since(startTime)
	if failed > 0 {
		fmt.Fprintf(w, "GOSS CRITICAL - Count: %d, Failed: %d, Skipped: %d, Duration: %.3fs\n", testCount, failed, skipped, duration.Seconds())
		return 2
	}
	fmt.Fprintf(w, "GOSS OK - Count: %d, Failed: %d, Skipped: %d, Duration: %.3fs\n", testCount, failed, skipped, duration.Seconds())
	return 0
}

func init() {
	RegisterOutputer("nagios", &Nagios{})
}
