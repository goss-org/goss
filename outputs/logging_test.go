package outputs

import (
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goss-org/goss/matchers"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func captureRecords(level slog.Level) (*slog.Logger, func() []map[string]any) {
	buf := &syncBuffer{}
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: util.ReplaceTraceLevel,
	})

	return slog.New(handler), func() []map[string]any {
		var out []map[string]any
		for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
			if line == "" {
				continue
			}
			var record map[string]any
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}
			out = append(out, record)
		}
		return out
	}
}

func recordsWithMessage(records []map[string]any, msg string) []map[string]any {
	var out []map[string]any
	for _, record := range records {
		if record["msg"] == msg {
			out = append(out, record)
		}
	}
	return out
}

// TestJSONSummaryRecord covers the json outputer's summary: one record per
// Output call, carrying the same document the writer received.
func TestJSONSummaryRecord(t *testing.T) {
	tests := map[string]struct {
		results    func() <-chan []resource.TestResult
		wantStatus string
		wantExit   int
	}{
		"passing": {results: passingResults, wantStatus: "ok", wantExit: 0},
		"failing": {results: failingResults, wantStatus: "fail", wantExit: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, records := captureRecords(slog.LevelDebug)
			written := &strings.Builder{}

			exit := Json{}.Output(written, tc.results(), util.OutputConfig{Logger: logger})
			require.Equal(t, tc.wantExit, exit)

			summaries := recordsWithMessage(records(), "validation summary")
			require.Len(t, summaries, 1, "exactly one summary per Output call")
			require.Equal(t, "DEBUG", summaries[0]["level"])
			require.Equal(t, tc.wantStatus, summaries[0]["status"])

			// Fprintln adds exactly one line delimiter beyond the payload.
			require.Equal(t, strings.TrimSuffix(written.String(), "\n"), summaries[0]["results_json"],
				"the record should carry the same document the writer received")
			require.NotEmpty(t, summaries[0]["results_json"])
		})
	}
}

// TestRspecishSummaryRecord covers rspecish, whose summary was counts and
// timings all along. They arrive as numbers instead of an escape-coloured
// string.
func TestRspecishSummaryRecord(t *testing.T) {
	tests := map[string]struct {
		results     func() <-chan []resource.TestResult
		wantStatus  string
		wantExit    int
		wantTotal   float64
		wantFailed  float64
		wantSkipped float64
	}{
		"passing": {results: passingResults, wantStatus: "ok", wantExit: 0, wantTotal: 1},
		"failing": {results: failingResults, wantStatus: "fail", wantExit: 1, wantTotal: 1, wantFailed: 1},
		"skipping": {
			results: skippedResults, wantStatus: "ok", wantExit: 0,
			wantTotal: 1, wantSkipped: 1,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, records := captureRecords(slog.LevelDebug)
			written := &strings.Builder{}

			exit := Rspecish{}.Output(written, tc.results(), util.OutputConfig{Logger: logger})
			require.Equal(t, tc.wantExit, exit)

			summaries := recordsWithMessage(records(), "validation summary")
			require.Len(t, summaries, 1)
			require.Equal(t, "DEBUG", summaries[0]["level"])
			require.Equal(t, tc.wantStatus, summaries[0]["status"])
			require.Equal(t, tc.wantTotal, summaries[0]["total"])
			require.Equal(t, tc.wantFailed, summaries[0]["failed"])
			require.Equal(t, tc.wantSkipped, summaries[0]["skipped"])
			require.IsType(t, float64(0), summaries[0]["duration_seconds"])

			require.NotContains(t, summaries[0], "results_json")
			for _, value := range summaries[0] {
				if s, ok := value.(string); ok {
					require.NotContains(t, s, "\x1b[", "no ANSI escapes belong in a record")
					require.NotContains(t, s, "Total Duration")
				}
			}
		})
	}
}

// TestValidationTraceRecords covers the two trace sites and the difference
// between them: json's records carry the numeric result and rspecish's do not.
//
// Every outcome is driven through both, because the two outputers decide it in
// different code: rspecish switches on the result, json used to treat anything
// that had not failed as a success.
func TestValidationTraceRecords(t *testing.T) {
	outcomes := map[string]struct {
		results func() <-chan []resource.TestResult
		outcome string
		result  int
	}{
		"success": {results: passingResults, outcome: "success", result: resource.SUCCESS},
		"fail":    {results: failingResults, outcome: "fail", result: resource.FAIL},
		"skip":    {results: skippedResults, outcome: "skip", result: resource.SKIP},
	}

	for name, tc := range outcomes {
		t.Run("json/"+name, func(t *testing.T) {
			logger, records := captureRecords(util.LevelTrace)

			Json{}.Output(io.Discard, tc.results(), util.OutputConfig{Logger: logger})

			traces := recordsWithMessage(records(), "validation result")
			require.Len(t, traces, 1)
			require.Equal(t, "TRACE", traces[0]["level"])
			require.Equal(t, tc.outcome, traces[0]["outcome"])
			require.Equal(t, "File", traces[0]["resource_type"])
			require.Equal(t, "exists", traces[0]["property"])
			require.Equal(t, float64(tc.result), traces[0]["result"],
				"the outcome and the numeric result must agree")
			require.IsType(t, float64(0), traces[0]["duration_seconds"])
			require.Contains(t, traces[0], "expected")
			require.Contains(t, traces[0], "actual")
		})

		t.Run("rspecish/"+name, func(t *testing.T) {
			logger, records := captureRecords(util.LevelTrace)

			Rspecish{}.Output(io.Discard, tc.results(), util.OutputConfig{Logger: logger})

			traces := recordsWithMessage(records(), "validation result")
			require.Len(t, traces, 1)
			require.Equal(t, "TRACE", traces[0]["level"])
			require.Equal(t, tc.outcome, traces[0]["outcome"])
			require.NotContains(t, traces[0], "result",
				"rspecish's trace records never carried the numeric result")
		})
	}
}

// TestSubjectPayloadsStayBelowDebug guards the level subject content lands at.
// The planted value is the kind of content that comes from the system under
// test, and it must not appear in a record an operator sees at INFO or above.
func TestSubjectPayloadsStayBelowDebug(t *testing.T) {
	const sentinel = "sentinel-from-the-subject"

	results := func() <-chan []resource.TestResult {
		c := make(chan []resource.TestResult, 1)
		now := time.Now()
		c <- []resource.TestResult{{
			Result: resource.FAIL, ResourceType: "Command", ResourceId: "probe",
			Property: "stdout", StartTime: now, EndTime: now,
			MatcherResult: matchers.MatcherResult{Actual: sentinel, Expected: "something else"},
		}}
		close(c)
		return c
	}

	t.Run("present at trace", func(t *testing.T) {
		logger, records := captureRecords(util.LevelTrace)

		Json{}.Output(io.Discard, results(), util.OutputConfig{Logger: logger})

		require.Contains(t, renderRecords(records()), sentinel)
	})

	t.Run("absent from info and above", func(t *testing.T) {
		logger, records := captureRecords(slog.LevelInfo)

		Json{}.Output(io.Discard, results(), util.OutputConfig{Logger: logger})

		require.NotContains(t, renderRecords(records()), sentinel)
	})
}

func renderRecords(records []map[string]any) string {
	var sb strings.Builder
	for _, record := range records {
		encoded, _ := json.Marshal(record)
		sb.Write(encoded)
	}
	return sb.String()
}

// TestOutputersWithoutALoggerAreSilent is the outputs/ half of the silence
// guarantee, and the reason util.OutputConfig stays usable as a zero value.
func TestOutputersWithoutALoggerAreSilent(t *testing.T) {
	for name, outputer := range map[string]Outputer{"json": Json{}, "rspecish": Rspecish{}} {
		t.Run(name, func(t *testing.T) {
			require.NotPanics(t, func() {
				outputer.Output(io.Discard, failingResults(), util.OutputConfig{})
			})
		})
	}
}

func passingResults() <-chan []resource.TestResult {
	now := time.Now()
	c := make(chan []resource.TestResult, 1)
	c <- []resource.TestResult{{
		Successful: true, Result: resource.SUCCESS, ResourceType: "File",
		ResourceId: "/tmp", Property: "exists", StartTime: now, EndTime: now,
	}}
	close(c)
	return c
}

func failingResults() <-chan []resource.TestResult {
	now := time.Now()
	c := make(chan []resource.TestResult, 1)
	c <- []resource.TestResult{{
		Successful: false, Result: resource.FAIL, ResourceType: "File",
		ResourceId: "/nope", Property: "exists", StartTime: now, EndTime: now,
	}}
	close(c)
	return c
}

func skippedResults() <-chan []resource.TestResult {
	now := time.Now()
	c := make(chan []resource.TestResult, 1)
	c <- []resource.TestResult{{
		Successful: true, Skipped: true, Result: resource.SKIP, ResourceType: "File",
		ResourceId: "/skipped", Property: "exists", StartTime: now, EndTime: now,
	}}
	close(c)
	return c
}
