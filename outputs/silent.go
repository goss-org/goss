package outputs

import (
	"io"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

type Silent struct{}

func (r Silent) ValidOptions() []*formatOption {
	return []*formatOption{}
}

func (r Silent) Output(w io.Writer, results <-chan []resource.TestResult,
	outConfig util.OutputConfig) (exitCode int) {

	var failed int
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.FAIL:
				failed++
			}
		}
	}

	if failed > 0 {
		return 1
	}
	return 0
}
