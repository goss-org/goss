package outputs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

type Json struct{}

func (r Json) ValidOptions() []*formatOption {
	return []*formatOption{
		{name: foPretty},
		{name: foSort},
	}
}

func (r Json) Output(w io.Writer, results <-chan []resource.TestResult,
	outConfig util.OutputConfig) (exitCode int) {

	var pretty bool = util.IsValueInList(foPretty, outConfig.FormatOptions)
	includeRaw := !util.IsValueInList(foExcludeRaw, outConfig.FormatOptions)

	sort := util.IsValueInList(foSort, outConfig.FormatOptions)
	results = getResults(results, sort)

	var startTime time.Time
	var endTime time.Time
	color.NoColor = true
	testCount := 0
	failed := 0
	skipped := 0
	var resultsOut []map[string]any
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if startTime.IsZero() || testResult.StartTime.Before(startTime) {
				startTime = testResult.StartTime
			}
			if endTime.IsZero() || testResult.EndTime.After(endTime) {
				endTime = testResult.EndTime
			}
			if testResult.Result == resource.FAIL {
				failed++
				logTrace("TRACE", "FAIL", testResult, true)
			} else {
				logTrace("TRACE", "SUCCESS", testResult, true)
			}
			if testResult.Skipped {
				skipped++
			}
			m := struct2map(testResult)
			m["successful"] = testResult.Result != resource.FAIL
			m["summary-line"] = humanizeResult(testResult, false, includeRaw)
			m["summary-line-compact"] = humanizeResult(testResult, true, includeRaw)
			m["duration"] = testResult.Duration.Nanoseconds()
			resultsOut = append(resultsOut, m)
			testCount++
		}
	}

	summary := make(map[string]any)
	duration := endTime.Sub(startTime)
	summary["test-count"] = testCount
	summary["failed-count"] = failed
	summary["skipped-count"] = skipped
	summary["total-duration"] = duration
	summary["summary-line"] = fmt.Sprintf("Count: %d, Failed: %d, Skipped: %d, Duration: %.3fs", testCount, failed, skipped, duration.Seconds())

	out := make(map[string]any)
	out["results"] = resultsOut
	out["summary"] = summary

	var j []byte
	if pretty {
		j, _ = json.MarshalIndent(out, "", "    ")
	} else {
		j, _ = json.Marshal(out)
	}

	resstr := string(j)
	fmt.Fprintln(w, resstr)

	if failed > 0 {
		log.Printf("[DEBUG] FAIL SUMMARY: %s", resstr)
		return 1
	}

	log.Printf("[DEBUG] OK SUMMARY: %s", resstr)
	return 0
}

func struct2map(i any) map[string]any {
	out := make(map[string]any)
	j, _ := json.Marshal(i)
	json.Unmarshal(j, &out)
	return out
}
