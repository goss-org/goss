package util

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// newJSONCapture returns a logger writing JSON records into buf at the given
// level. slog serialises handler writes internally, so a plain buffer is safe
// even when records are emitted from several goroutines.
func newJSONCapture(level slog.Leveler, addSource bool) (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		AddSource:   addSource,
		Level:       level,
		ReplaceAttr: ReplaceTraceLevel,
	})
	return slog.New(handler), buf
}

// decodeRecords parses the newline-delimited JSON records written to buf.
func decodeRecords(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var out []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &rec), "record %q should be JSON", line)
		out = append(out, rec)
	}
	return out
}

// TestReplaceTraceLevelRendersTrace pins the rendering: LevelTrace must appear
// as TRACE when the exported hook is installed directly as a handler's
// ReplaceAttr. Without the hook slog renders the level as "DEBUG-4".
func TestReplaceTraceLevelRendersTrace(t *testing.T) {
	t.Parallel()

	t.Run("text handler", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		// Used directly as the ReplaceAttr value: this assignment is part of
		// the assertion, since the hook must carry that exact signature.
		logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level:       LevelTrace,
			ReplaceAttr: ReplaceTraceLevel,
		}))

		Trace(logger, "traced")

		got := buf.String()
		require.Contains(t, got, "level=TRACE")
		require.NotContains(t, got, "DEBUG-4")
	})

	t.Run("json handler", func(t *testing.T) {
		t.Parallel()

		logger, buf := newJSONCapture(LevelTrace, false)

		Trace(logger, "traced")

		records := decodeRecords(t, buf)
		require.Len(t, records, 1)
		require.Equal(t, "TRACE", records[0]["level"])
	})
}

// TestReplaceTraceLevelLeavesOtherAttrsAlone pins the hook's blast radius: it
// rewrites the built-in level attribute for LevelTrace and nothing else.
func TestReplaceTraceLevelLeavesOtherAttrsAlone(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		groups []string
		attr   slog.Attr
		want   slog.Value
	}{
		"debug level is untouched": {
			attr: slog.Any(slog.LevelKey, slog.LevelDebug),
			want: slog.AnyValue(slog.LevelDebug),
		},
		"error level is untouched": {
			attr: slog.Any(slog.LevelKey, slog.LevelError),
			want: slog.AnyValue(slog.LevelError),
		},
		"a user attribute named level is untouched": {
			attr: slog.String(slog.LevelKey, "whatever"),
			want: slog.StringValue("whatever"),
		},
		"a grouped level attribute is untouched": {
			groups: []string{"nested"},
			attr:   slog.Any(slog.LevelKey, LevelTrace),
			want:   slog.AnyValue(LevelTrace),
		},
		"a differently keyed trace level is untouched": {
			attr: slog.Any("configured_level", LevelTrace),
			want: slog.AnyValue(LevelTrace),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ReplaceTraceLevel(tc.groups, tc.attr)
			require.Equal(t, tc.attr.Key, got.Key)
			require.True(t, got.Value.Equal(tc.want),
				"expected %v got %v", tc.want, got.Value)
		})
	}
}

// TestTraceWrappersReportCallerSource checks that the source position on a
// record emitted through the wrappers must be the calling site, not the
// wrapper's own file.
func TestTraceWrappersReportCallerSource(t *testing.T) {
	t.Parallel()

	t.Run("Trace", func(t *testing.T) {
		t.Parallel()

		logger, buf := newJSONCapture(LevelTrace, true)

		_, wantFile, callerLine, ok := runtime.Caller(0)
		require.True(t, ok)
		Trace(logger, "traced") // two lines below runtime.Caller

		assertSource(t, buf, wantFile, callerLine+2)
	})

	t.Run("TraceContext", func(t *testing.T) {
		t.Parallel()

		logger, buf := newJSONCapture(LevelTrace, true)

		_, wantFile, callerLine, ok := runtime.Caller(0)
		require.True(t, ok)
		TraceContext(context.Background(), logger, "traced") // two lines below runtime.Caller

		assertSource(t, buf, wantFile, callerLine+2)
	})
}

func assertSource(t *testing.T, buf *bytes.Buffer, wantFile string, wantLine int) {
	t.Helper()

	records := decodeRecords(t, buf)
	require.Len(t, records, 1)

	source, ok := records[0]["source"].(map[string]any)
	require.True(t, ok, "record should carry a source group: %v", records[0])
	require.Equal(t, filepath.Base(wantFile), filepath.Base(source["file"].(string)))
	require.Equal(t, float64(wantLine), source["line"])
}

// TestTraceWrappersAreNilSafe covers the guard for the wrappers themselves: a
// nil logger must not panic, since util.Config.Logger is nil for the documented
// composite-literal embedding pattern.
func TestTraceWrappersAreNilSafe(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		Trace(nil, "traced", "key", "value")
		TraceContext(context.Background(), nil, "traced", "key", "value")
		//nolint:staticcheck // a nil context is exactly what is under test here
		TraceContext(nil, nil, "traced")
	})
}

// TestTraceWrappersHonourEnabled is the control for the caller-source test
// above, and the evidence that TRACE records are suppressed from DEBUG upwards.
func TestTraceWrappersHonourEnabled(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		level   slog.Level
		visible bool
	}{
		"trace emits": {level: LevelTrace, visible: true},
		"debug mutes": {level: slog.LevelDebug, visible: false},
		"info mutes":  {level: slog.LevelInfo, visible: false},
		"warn mutes":  {level: slog.LevelWarn, visible: false},
		"error mutes": {level: slog.LevelError, visible: false},
		"below trace": {level: LevelTrace - 1, visible: true},
		"above trace": {level: LevelTrace + 1, visible: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logger, buf := newJSONCapture(tc.level, false)

			Trace(logger, "traced", "sentinel", "trace-sentinel")
			TraceContext(context.Background(), logger, "traced", "sentinel", "context-sentinel")

			records := decodeRecords(t, buf)
			if tc.visible {
				require.Len(t, records, 2)
				require.Equal(t, "trace-sentinel", records[0]["sentinel"])
				require.Equal(t, "context-sentinel", records[1]["sentinel"])
			} else {
				require.Empty(t, records)
			}
		})
	}
}

// TestLevelTraceIsBelowDebug pins the constant's relationship to the standard
// levels, which is what makes --log-level DEBUG mute TRACE records.
func TestLevelTraceIsBelowDebug(t *testing.T) {
	t.Parallel()

	require.Less(t, LevelTrace, slog.LevelDebug)
	require.Equal(t, slog.Level(-8), LevelTrace)
}

// TestLoggerOrDiscard pins the guard every logger derivation goes through, so
// it must return a usable logger for nil and pass a real logger through
// untouched.
func TestLoggerOrDiscard(t *testing.T) {
	t.Parallel()

	t.Run("nil yields a silent usable logger", func(t *testing.T) {
		t.Parallel()

		got := LoggerOrDiscard(nil)
		require.NotNil(t, got)
		require.NotPanics(t, func() {
			got.Debug("debug")
			got.Info("info")
			got.Error("error")
			got.With("key", "value").Warn("warn")
			require.False(t, got.Enabled(context.Background(), slog.LevelError))
		})
	})

	t.Run("a supplied logger is returned unchanged", func(t *testing.T) {
		t.Parallel()

		logger, buf := newJSONCapture(slog.LevelInfo, false)
		require.Same(t, logger, LoggerOrDiscard(logger))

		LoggerOrDiscard(logger).Info("passed through")
		require.Contains(t, buf.String(), "passed through")
	})
}
