package outputs

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	outputer := &Prometheus{}
	injectedResults := []resource.TestResult{
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
		{
			ResourceType: "Command",
			Duration:     10 * time.Millisecond,
			Result:       resource.FAIL,
		},
		{
			ResourceType: "File",
			Duration:     10 * time.Millisecond,
			Result:       resource.SKIP,
		},
	}

	exitCode := outputer.Output(buf, makeResults(injectedResults...), util.OutputConfig{})

	assert.Equal(t, 0, exitCode)
	output := buf.String()
	t.Logf(output)
	assert.Contains(t, output, `goss_tests_outcomes_duration_milliseconds{outcome="pass",type="command"} 20`)
	assert.Contains(t, output, `goss_tests_outcomes_duration_milliseconds{outcome="fail",type="command"} 10`)
	assert.Contains(t, output, `goss_tests_outcomes_duration_milliseconds{outcome="skip",type="command"} 0`)
	assert.Contains(t, output, `goss_tests_outcomes_duration_milliseconds{outcome="pass",type="file"} 0`)
	assert.Contains(t, output, `goss_tests_outcomes_duration_milliseconds{outcome="fail",type="file"} 0`)
	assert.Contains(t, output, `goss_tests_outcomes_duration_milliseconds{outcome="skip",type="file"} 10`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="pass",type="command"} 2`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="fail",type="command"} 1`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="skip",type="command"} 0`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="pass",type="file"} 0`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="fail",type="file"} 0`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="skip",type="file"} 1`)
	assert.Contains(t, output, `goss_tests_run_duration_milliseconds{outcome="skip"} 60000`)
	assert.Contains(t, output, `goss_tests_run_outcomes_total{outcome="skip"} 1`)
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
