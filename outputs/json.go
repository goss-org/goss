package outputs

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aelsabbahy/goss/resource"
)

type Json struct{}

func (r Json) Output(results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	testCount := 0
	failed := 0
	var resultsOut []resource.TestResult
	for resultGroup := range results {
		for _, testResult := range resultGroup {
			if !testResult.Successful {
				failed++
			}
			resultsOut = append(resultsOut, testResult)
			testCount++
		}
	}

	summary := make(map[string]interface{})
	summary["test-count"] = testCount
	summary["failed-count"] = failed
	summary["total-duration"] = time.Now().Sub(startTime)

	out := make(map[string]interface{})
	out["results"] = resultsOut
	out["summary"] = summary

	j, _ := json.MarshalIndent(out, "", "    ")
	fmt.Println(string(j))

	if failed > 0 {
		return 1
	}

	return 0
}

func init() {
	RegisterOutputer("json", &Json{})
}
