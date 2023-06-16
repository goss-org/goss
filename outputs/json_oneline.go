package outputs

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

type JsonOneline struct{}

func (r JsonOneline) ValidOptions() []*formatOption {
	return []*formatOption{}
}

func (r JsonOneline) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {

	color.NoColor = true
	testCount := 0
	failed := 0
	var resultsOut []map[string]any
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if !testResult.Successful {
				failed++
			}
			m := struct2map(testResult)
			m["summary-line"] = humanizeResult(testResult)
			m["duration"] = int64(m["duration"].(float64))
			resultsOut = append(resultsOut, m)
			testCount++
		}
	}

	summary := make(map[string]any)
	duration := time.Since(startTime)
	summary["test-count"] = testCount
	summary["failed-count"] = failed
	summary["total-duration"] = duration
	summary["summary-line"] = fmt.Sprintf("Count: %d, Failed: %d, Duration: %.3fs", testCount, failed, duration.Seconds())

	out := make(map[string]any)
	out["results"] = resultsOut
	out["summary"] = summary

	j, _ := json.Marshal(out)
	fmt.Fprintln(w, string(j))

	if failed > 0 {
		return 1
	}

	return 0
}
