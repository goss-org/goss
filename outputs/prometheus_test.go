package outputs

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/aelsabbahy/goss/resource"

	"github.com/aelsabbahy/goss/util"
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

	exitCode := outputer.Output(buf, makeResults(injectedResults...), time.Now().Add(-1*time.Minute), util.OutputConfig{})

	assert.Equal(t, 0, exitCode)
	output := buf.String()
	t.Logf(output)
	assert.Contains(t, output, `goss_tests_outcomes_duration_seconds{outcome="pass",type="command"} 0.02`)
	assert.Contains(t, output, `goss_tests_outcomes_duration_seconds{outcome="fail",type="command"} 0.01`)
	assert.Contains(t, output, `goss_tests_outcomes_duration_seconds{outcome="skip",type="file"} 0.01`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="pass",type="command"} 2`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="fail",type="command"} 1`)
	assert.Contains(t, output, `goss_tests_outcomes_total{outcome="skip",type="file"} 1`)
	assert.Contains(t, output, "goss_tests_run_duration_seconds 60")
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
