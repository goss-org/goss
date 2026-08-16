package goss

import (
	"encoding/json"
	"log"
	"log/slog"
	"strings"
	"testing"

	"github.com/goss-org/goss/util"
)

// captureRecords returns a logger writing JSON records, and a function decoding
// the ones emitted so far.
func captureRecords(level slog.Level) (*slog.Logger, func() []map[string]any) {
	buf := &syncBuffer{}
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: util.ReplaceTraceLevel,
	})

	return slog.New(handler), func() []map[string]any { return decodeJSONRecords(buf.String()) }
}

func decodeJSONRecords(raw string) []map[string]any {
	var out []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(raw), "\n") {
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

// recordsWithMessage picks out the records carrying one constant message, which
// is how a structured record is identified now that messages hold no data.
func recordsWithMessage(records []map[string]any, msg string) []map[string]any {
	var out []map[string]any
	for _, record := range records {
		if record["msg"] == msg {
			out = append(out, record)
		}
	}
	return out
}

// captureGlobalLogger redirects both process-global logging APIs for a test
// and returns their raw output. Tests using it must remain serial.
func captureGlobalLogger(t *testing.T) func() string {
	t.Helper()

	buf := &syncBuffer{}
	originalLogger := slog.Default()
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: util.LevelTrace,
	})))
	log.SetOutput(buf)
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
	})

	return buf.String
}
