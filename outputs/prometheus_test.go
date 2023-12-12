package outputs

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusOutput(t *testing.T) {
	testCases := map[string]struct {
		results         []resource.TestResult
		expectedMetrics []string
	}{
		"all-success-single-type": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 20`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 2`,
				`goss_tests_run_duration_milliseconds{outcome="pass"}`,
				`goss_tests_run_outcomes_total{outcome="pass"} 1`,
			},
		},
		"all-skip-single-type": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 20`,
				`goss_tests_outcomes_total{outcome="skip",type="command"} 2`,
				`goss_tests_run_duration_milliseconds{outcome="skip"}`,
				`goss_tests_run_outcomes_total{outcome="skip"} 1`,
			},
		},
		"all-fail-single-type": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 20`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 2`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"all-unknown-single-type": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="command"} 20`,
				`goss_tests_outcomes_total{outcome="unknown",type="command"} 2`,
				`goss_tests_run_duration_milliseconds{outcome="unknown"}`,
				`goss_tests_run_outcomes_total{outcome="unknown"} 1`,
			},
		},
		"all-success-multiple-types": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
				{
					ResourceType: "File",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="file"} 10`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="pass",type="file"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="pass"}`,
				`goss_tests_run_outcomes_total{outcome="pass"} 1`,
			},
		},
		"various-results-single-type": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="skip",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="unknown",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"various-results-multiple-types": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
				{
					ResourceType: "File",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
				{
					ResourceType: "File",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="file"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="file"} 10`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="skip",type="file"} 1`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="unknown",type="file"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"unknown-skip": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="skip",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="unknown",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="unknown"}`,
				`goss_tests_run_outcomes_total{outcome="unknown"} 1`,
			},
		},
		"unknown-fail": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="unknown",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"unknown-success": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="unknown",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="unknown"}`,
				`goss_tests_run_outcomes_total{outcome="unknown"} 1`,
			},
		},
		"skip-unknown": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="skip",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="unknown",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="unknown"}`,
				`goss_tests_run_outcomes_total{outcome="unknown"} 1`,
			},
		},
		"skip-fail": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="skip",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"skip-success": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="skip",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="pass"}`,
				`goss_tests_run_outcomes_total{outcome="pass"} 1`,
			},
		},
		"fail-unknown": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="unknown",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"fail-skip": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="skip",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"fail-success": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"success-unknown": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.UNKNOWN,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="unknown",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="unknown",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="unknown"}`,
				`goss_tests_run_outcomes_total{outcome="unknown"} 1`,
			},
		},
		"success-skip": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SKIP,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="skip",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="pass"}`,
				`goss_tests_run_outcomes_total{outcome="pass"} 1`,
			},
		},
		"success-fail": {
			results: []resource.TestResult{
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.SUCCESS,
				},
				{
					ResourceType: "Command",
					Duration:     10 * time.Millisecond,
					Result:       resource.FAIL,
				},
			},
			expectedMetrics: []string{
				`goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 10`,
				`goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`,
				`goss_tests_outcomes_total{outcome="pass",type="command"} 1`,
				`goss_tests_outcomes_total{outcome="fail",type="command"} 1`,
				`goss_tests_run_duration_milliseconds{outcome="fail"}`,
				`goss_tests_run_outcomes_total{outcome="fail"} 1`,
			},
		},
		"no-results": {
			results: []resource.TestResult{},
			expectedMetrics: []string{
				`goss_tests_run_duration_milliseconds{outcome="unknown"}`,
				`goss_tests_run_outcomes_total{outcome="unknown"} 1`,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			outputer := &Prometheus{}

			defer resetMetrics()

			exitCode := outputer.Output(buf, makeResults(testCase.results...), util.OutputConfig{})
			assert.Equal(t, 0, exitCode)

			output := buf.String()
			t.Logf(output)
			for _, metric := range testCase.expectedMetrics {
				assert.Contains(t, output, metric)
			}
		})
	}
}

func makeResults(results ...resource.TestResult) <-chan []resource.TestResult {
	out := make(chan []resource.TestResult)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		out <- append([]resource.TestResult{}, results...)
	}()

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func resetMetrics() {
	testOutcomes.Reset()
	testDurations.Reset()
	runOutcomes.Reset()
	runDuration.Reset()
}

func TestCanChangeOverallOutcome(t *testing.T) {
	testCases := map[string]map[string]bool{
		resource.OutcomePass: {
			resource.OutcomePass:    true,
			resource.OutcomeSkip:    false,
			resource.OutcomeFail:    true,
			resource.OutcomeUnknown: true,
		},
		resource.OutcomeSkip: {
			resource.OutcomePass:    true,
			resource.OutcomeSkip:    true,
			resource.OutcomeFail:    true,
			resource.OutcomeUnknown: true,
		},
		resource.OutcomeFail: {
			resource.OutcomePass:    false,
			resource.OutcomeSkip:    false,
			resource.OutcomeFail:    false,
			resource.OutcomeUnknown: false,
		},
		resource.OutcomeUnknown: {
			resource.OutcomePass:    false,
			resource.OutcomeSkip:    false,
			resource.OutcomeFail:    true,
			resource.OutcomeUnknown: false,
		},
	}
	for current, expectations := range testCases {
		for result, canChange := range expectations {
			t.Run(fmt.Sprintf("%s/%s", current, result), func(t *testing.T) {
				assert.Equalf(t, canChange, canChangeOverallOutcome(current, result), "canChangeOverallOutcome(%v, %v)", current, result)
			})
		}
	}
}
