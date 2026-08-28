package outputs

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

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

	logger := util.LoggerOrDiscard(outConfig.Logger)

	var pretty = util.IsValueInList(foPretty, outConfig.FormatOptions)
	includeRaw := !util.IsValueInList(foExcludeRaw, outConfig.FormatOptions)

	sort := util.IsValueInList(foSort, outConfig.FormatOptions)
	results = getResults(results, sort)

	var startTime time.Time
	var endTime time.Time
	SetNoColor(true)
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
			}
			if testResult.Skipped {
				skipped++
			}
			// The counters above are unchanged. The outcome is not: the old
			// message said "SUCCESS" for a skipped result, which was survivable
			// in prose next to the result number and is not survivable as an
			// enum attribute called outcome.
			logTrace(logger, outcomeFor(testResult), testResult, true)
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

	// One record per Output call, carrying the same document written to w. The
	// two summary sites this replaces differed only in the word they put in
	// front of it, which is now an attribute.
	if failed > 0 {
		logger.Debug("validation summary", "status", statusFail, "results_json", resstr)
		return 1
	}

	logger.Debug("validation summary", "status", statusOK, "results_json", resstr)
	return 0
}

func struct2map(i any) map[string]any {
	out := make(map[string]any)
	j, _ := json.Marshal(i)
	json.Unmarshal(j, &out)
	return out
}
