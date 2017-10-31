package outputs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/fatih/color"
)

type Json struct {
	Report *url.URL
}

func (r *Json) SetReportURL(stringified string) error {
	u, err := url.Parse(stringified)
	if err != nil {
		return err
	}

	r.Report = u

	return nil
}

func (r *Json) Output(w io.Writer, results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	color.NoColor = true

	out, failed := makeMap(results, startTime)

	j, _ := json.MarshalIndent(out, "", "    ")
	fmt.Fprintln(w, string(j))

	if r.Report != nil {
		if err := postReport(j, r.Report); err != nil {
			fmt.Errorf("errors sending report: %s", err.Error())
		}
	}

	if failed > 0 {
		return 1
	}

	return 0
}

func init() {
	RegisterOutputer("json", &Json{})
}

func makeMap(results <-chan []resource.TestResult, startTime time.Time) (map[string]interface{}, int) {
	var (
		testCount  = 0
		failed     = 0
		resultsOut []map[string]interface{}
	)

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

	summary := make(map[string]interface{})
	duration := time.Since(startTime)
	summary["test-count"] = testCount
	summary["failed-count"] = failed
	summary["total-duration"] = duration
	summary["summary-line"] = fmt.Sprintf("Count: %d, Failed: %d, Duration: %.3fs", testCount, failed, duration.Seconds())

	hostname, _ := os.Hostname()
	out := make(map[string]interface{})
	out["hostname"] = hostname
	out["results"] = resultsOut
	out["summary"] = summary

	return out, failed
}

func struct2map(i interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	j, _ := json.Marshal(i)
	json.Unmarshal(j, &out)
	return out
}
