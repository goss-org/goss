package outputs

import (
	"fmt"
	"io"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
)

type Bench struct{}

func (r Bench) Output(w io.Writer, results <-chan []resource.TestResult, startTime time.Time, outConfig util.OutputConfig) (exitCode int) {
	includeRaw := util.IsValueInList("include_raw", outConfig.FormatOptions)
	var testCount, skipped, failed int
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			fmt.Fprintf(w, "%v %s\n", testResult.Duration, humanizeResult(testResult, true, includeRaw))
			switch testResult.Result {
			case resource.SKIP:
				skipped++
			case resource.FAIL:
				failed++
			}
			testCount++
		}
	}

	fmt.Fprint(w, summary(startTime, testCount, failed, skipped))
	if failed > 0 {
		return 1
	}
	return 0
}

func init() {
	RegisterOutputer("bench", &Bench{}, []string{})
}
