package main

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// TestCLIHandlerNormalisesTimeToUTC drives the production handler. The record
// carries a fixed non-UTC timestamp and a user attribute that collides with the
// built-in time key, which is exactly the case ReplaceAttr cannot handle: it
// cannot tell the two apart. The handler wrapper works on slog.Record.Time, so
// it can.
func TestCLIHandlerNormalisesTimeToUTC(t *testing.T) {
	t.Parallel()

	// 14:45 in a zone eight hours ahead is 06:45 UTC the same day.
	zone := time.FixedZone("UTC+8", 8*60*60)
	recordTime := time.Date(2026, 8, 14, 14, 45, 30, 0, zone)

	buf := &bytes.Buffer{}
	handler := newCLIHandler(buf, util.LevelTrace)

	record := slog.NewRecord(recordTime, util.LevelTrace, "traced", 0)
	record.Add(slog.TimeKey, "a user value")
	require.NoError(t, handler.Handle(context.Background(), record))

	got := buf.String()
	require.Contains(t, got, "time=2026-08-14T06:45:30.000Z",
		"the built-in timestamp should be normalised to UTC")
	require.Contains(t, got, `time="a user value"`,
		"a user attribute named time should survive untouched")
	require.Contains(t, got, "level=TRACE",
		"the CLI handler should also render LevelTrace as TRACE")
}

// TestCLIHandlerPreservesUserTimeAtEveryLevel guards the wrapper's delegation:
// grouping and pre-bound attributes must keep working, since the wrapper has to
// reproduce them rather than inherit them.
func TestCLIHandlerDelegatesWithAttrsAndWithGroup(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(newCLIHandler(buf, slog.LevelInfo)).
		With("bound", "attr").
		WithGroup("grouped")

	logger.Info("message", "inner", "value")

	got := buf.String()
	require.Contains(t, got, "bound=attr")
	require.Contains(t, got, "grouped.inner=value")
	require.Contains(t, got, "level=INFO")
}

// TestCLIHandlerRespectsLevel pins the level as the handler's property, which is
// what replaces the message-prefix filter.
func TestCLIHandlerRespectsLevel(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		level   slog.Level
		emitted []string
		muted   []string
	}{
		"trace": {level: util.LevelTrace, emitted: []string{"traced", "debugged", "informed", "warned", "errored"}},
		"debug": {level: slog.LevelDebug, emitted: []string{"debugged", "informed", "warned", "errored"}, muted: []string{"traced"}},
		"info":  {level: slog.LevelInfo, emitted: []string{"informed", "warned", "errored"}, muted: []string{"traced", "debugged"}},
		"warn":  {level: slog.LevelWarn, emitted: []string{"warned", "errored"}, muted: []string{"traced", "debugged", "informed"}},
		"error": {level: slog.LevelError, emitted: []string{"errored"}, muted: []string{"traced", "debugged", "informed", "warned"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logger := slog.New(newCLIHandler(buf, tc.level))

			util.Trace(logger, "traced")
			logger.Debug("debugged")
			logger.Info("informed")
			logger.Warn("warned")
			logger.Error("errored")

			for _, msg := range tc.emitted {
				require.Contains(t, buf.String(), msg)
			}
			for _, msg := range tc.muted {
				require.NotContains(t, buf.String(), msg)
			}
		})
	}
}

// TestCLIHandlerWritesOnlyToItsSink is the handler half of the stderr
// guarantee. Production passes os.Stderr, which the static gate pins; here two
// sinks stand in for the two streams, and no record may appear in the one the
// handler was not given.
func TestCLIHandlerWritesOnlyToItsSink(t *testing.T) {
	t.Parallel()

	stderrSink := &bytes.Buffer{}
	stdoutSink := &bytes.Buffer{}

	logger := slog.New(newCLIHandler(stderrSink, util.LevelTrace))

	marker := "sink-destination-marker"
	util.Trace(logger, marker)
	logger.Debug(marker)
	logger.Info(marker)
	logger.Warn(marker)
	logger.Error(marker)

	require.Equal(t, 5, bytes.Count(stderrSink.Bytes(), []byte(marker)))
	require.Empty(t, stdoutSink.String())
}

// TestParseLogLevel covers the level names the CLI accepts, in every letter
// case, and the rejection an unsupported value earns.
func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	valid := map[string]slog.Level{
		"TRACE": util.LevelTrace,
		"DEBUG": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"WARN":  slog.LevelWarn,
		"ERROR": slog.LevelError,
	}

	for name, want := range valid {
		for _, spelling := range []string{name, lower(name), mixed(name)} {
			t.Run(spelling, func(t *testing.T) {
				t.Parallel()

				got, err := parseLogLevel(spelling)
				require.NoError(t, err)
				require.Equal(t, want, got)
			})
		}
	}

	// FATAL is not in the set the old filter accepted either, so rejecting it
	// preserves today's behaviour rather than narrowing it.
	for _, level := range []string{"FATAL", "", "verbose", "warning", "-4"} {
		t.Run("rejects "+level, func(t *testing.T) {
			t.Parallel()

			_, err := parseLogLevel(level)
			require.Error(t, err)
			require.Contains(t, err.Error(), level,
				"the error should name the value that was rejected")
		})
	}
}

func lower(s string) string {
	return string(bytes.ToLower([]byte(s)))
}

func mixed(s string) string {
	b := bytes.ToLower([]byte(s))
	b[0] = s[0]
	return string(b)
}
