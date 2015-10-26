package outputs

import (
	"encoding/json"
	"fmt"

	"github.com/aelsabbahy/goss/resource"
)

type Json struct{}

func (r Json) Output(results <-chan []resource.TestResult) (hasFail bool) {
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
	summary["test_count"] = testCount
	summary["failed_count"] = failed

	out := make(map[string]interface{})
	out["results"] = resultsOut
	out["summary"] = summary

	j, _ := json.MarshalIndent(out, "", "    ")
	fmt.Println(string(j))

	if failed > 0 {
		return true
	}

	return false
}

func init() {
	RegisterOutputer("json", &Json{})
}
