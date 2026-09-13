package outputs

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

func TestIsValidFormat(t *testing.T) {
	if IsValidFormat("ne") {
		t.Fatal("'ne' should not be a valid output format")
	}

	if !IsValidFormat("json") {
		t.Fatal("'json' should be a valid output format")
	}
}

func TestOutputers(t *testing.T) {
	list := Outputers()
	assert.NotEmpty(t, list)
}

func TestGetOutputer(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := GetOutputer("rspecish")
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("not-valid", func(t *testing.T) {
		got, err := GetOutputer("gibberish")
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestOutputFormatOptions(t *testing.T) {
	list := FormatOptions()
	assert.NotEmpty(t, list)

	assert.Contains(t, list, foPerfData)
	assert.Contains(t, list, foPretty)
	assert.Contains(t, list, foVerbose)
	assert.Len(t, list, 4)
}

func TestOptionsRegistration(t *testing.T) {
	registeredOutputs := Outputers()
	assert.Contains(t, registeredOutputs, "documentation")
	assert.Contains(t, registeredOutputs, "json")
	assert.Contains(t, registeredOutputs, "junit")
	assert.Contains(t, registeredOutputs, "nagios")
	assert.Contains(t, registeredOutputs, "prometheus")
	assert.Contains(t, registeredOutputs, "rspecish")
	assert.Contains(t, registeredOutputs, "silent")
	assert.Contains(t, registeredOutputs, "structured")
	assert.Contains(t, registeredOutputs, "tap")
}

// TestOutputExitCodes pins what every outputer returns, because serve turns a
// non-zero code into a 503 at /healthz. Prometheus was missing this once (#992)
// and structured was missing it too, so all of them are pinned at once rather
// than one per incident.
func TestOutputExitCodes(t *testing.T) {
	now := time.Now()
	pass := resource.TestResult{
		Result: resource.SUCCESS, ResourceType: "File", ResourceId: "/tmp",
		Property: "exists", StartTime: now, EndTime: now,
	}
	// a skip is not a failure, so it must not move the exit code on its own
	skip := resource.TestResult{
		Result: resource.SKIP, ResourceType: "File", ResourceId: "/skip",
		Property: "exists", Skipped: true, StartTime: now, EndTime: now,
	}
	fail := resource.TestResult{
		Result: resource.FAIL, ResourceType: "File", ResourceId: "/nope",
		Property: "exists", StartTime: now, EndTime: now,
	}

	cases := map[string]struct {
		outputer Outputer
		failing  int
	}{
		"documentation": {Documentation{}, 1},
		"json":          {Json{}, 1},
		"junit":         {JUnit{}, 1},
		"nagios":        {Nagios{}, 2},
		"prometheus":    {Prometheus{}, 1},
		"rspecish":      {Rspecish{}, 1},
		"silent":        {Silent{}, 1},
		"structured":    {Structured{}, 1},
		"tap":           {Tap{}, 1},
	}

	// a new outputer has to make a deliberate choice here rather than default to 0
	assert.Len(t, cases, len(Outputers()))

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// the prometheus outputer accumulates into package-level counters
			defer resetMetrics()

			assert.Equal(t, 0,
				tc.outputer.Output(io.Discard, makeResults(pass, skip), util.OutputConfig{}),
				"a suite with no failures must exit 0")
			assert.Equal(t, tc.failing,
				tc.outputer.Output(io.Discard, makeResults(pass, skip, fail), util.OutputConfig{}),
				"a suite containing a failure must not exit 0")
		})
	}
}
