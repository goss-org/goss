package outputs

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
)

type NagiosVerbose struct{}

func (r NagiosVerbose) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {

	var testCount, failed, skipped int

	var summary map[int]string
	summary = make(map[int]string)

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.FAIL:
				summary[failed] = "Fail " + strconv.Itoa(failed+1) + " - " + humanizeResult2(testResult) + "\n"
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
		for i := 0; i < failed; i++ {
			fmt.Fprintf(w, "%s", summary[i])
		}
		return 2
	}
	fmt.Fprintf(w, "GOSS OK - Count: %d, Failed: %d, Skipped: %d, Duration: %.3fs\n", testCount, failed, skipped, duration.Seconds())
	return 0
}

func init() {
	RegisterOutputer("nagios_verbose", &NagiosVerbose{}, []string{})
}
