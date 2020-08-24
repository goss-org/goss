package outputs

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
)

// Structured is a output formatter that logs into a StructuredOutput structure
type Structured struct{}

// StructuredTestResult is an individual test result with additional human friendly summary
type StructuredTestResult struct {
	resource.TestResult
	SummaryLine string `json:"summary-line"`
}

// StructureTestSummary holds summary information about a test run
type StructureTestSummary struct {
	TestCount     int           `json:"test-count"`
	Failed        int           `json:"failed-count"`
	TotalDuration time.Duration `json:"total-duration"`
}

// StructuredOutput is the full output structure for the structured output format
type StructuredOutput struct {
	Results     []StructuredTestResult `json:"results"`
	Summary     StructureTestSummary   `json:"summary"`
	SummaryLine string                 `json:"summary-line"`
}

// String represents human friendly representation of the test summary
func (s *StructureTestSummary) String() string {
	return fmt.Sprintf("Count: %d, Failed: %d, Duration: %.3fs", s.TestCount, s.Failed, s.TotalDuration.Seconds())
}

// Output processes output from tests into StructuredOutput written to w as a string
func (r Structured) Output(w io.Writer, results <-chan []resource.TestResult, startTime time.Time, outConfig util.OutputConfig) (exitCode int) {
	result := &StructuredOutput{
		Results: []StructuredTestResult{},
		Summary: StructureTestSummary{},
	}

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			r := StructuredTestResult{
				TestResult:  testResult,
				SummaryLine: humanizeResult(testResult),
			}

			if !testResult.Successful {
				result.Summary.Failed++
			}

			result.Summary.TestCount++

			result.Results = append(result.Results, r)
		}
	}

	result.Summary.TotalDuration = time.Since(startTime)
	result.SummaryLine = result.Summary.String()

	var j []byte

	if util.IsValueInList("pretty", outConfig.FormatOptions) {
		j, _ = json.MarshalIndent(result, "", "  ")
	} else {
		j, _ = json.Marshal(result)
	}

	fmt.Fprintln(w, string(j))

	if result.Summary.Failed > 0 {
		return 1
	}

	return 0
}

func init() {
	RegisterOutputer("structured", &Structured{}, []string{})
}
