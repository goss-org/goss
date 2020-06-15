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
	failed := false
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			fmt.Fprintf(w, "%v %s\n", testResult.Duration, humanizeResult(testResult, true, includeRaw))
			if testResult.Result == resource.SUCCESS {
				failed = true
			}
		}
	}

	if failed {
		return 1
	}
	return 0
}

func init() {
	RegisterOutputer("bench", &Bench{}, []string{})
}
