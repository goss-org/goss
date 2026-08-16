package system

import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"testing"

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
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})

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

// TestCommandOutputReachesTheInjectedLogger drives a real command. The record
// carries both identifiers, which matters because a resource's id and the
// command it executes are frequently not the same string.
func TestCommandOutputReachesTheInjectedLogger(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the probe command is a POSIX shell one")
	}

	logger, records := captureRecords(slog.LevelDebug)
	sys := New("", WithLogger(logger))

	// The id is deliberately not the command, so a record that confused the two
	// would be visible here.
	ctx := context.WithValue(context.Background(), CommandIDKey, "a-resource-id")
	config, err := util.NewConfig()
	require.NoError(t, err)
	config.Timeout = 10_000_000_000

	command := sys.NewCommand(ctx, `echo stdout-marker; echo stderr-marker >&2`, sys, *config)
	status, err := command.ExitStatus()
	require.NoError(t, err)
	require.Equal(t, 0, status)

	var stdout, stderr []map[string]any
	for _, record := range records() {
		require.Equal(t, "command output", record["msg"])
		require.Equal(t, "DEBUG", record["level"])
		require.Equal(t, `echo stdout-marker; echo stderr-marker >&2`, record["command"],
			"the executed command identifies the record")
		require.Equal(t, "a-resource-id", record["resource_id"],
			"so does the id the gossfile gave the resource")
		switch record["stream"] {
		case "stdout":
			stdout = append(stdout, record)
		case "stderr":
			stderr = append(stderr, record)
		default:
			t.Fatalf("unexpected stream %v", record["stream"])
		}
	}

	require.Len(t, stdout, 2, "the planted line plus the trailing empty one")
	require.Equal(t, "stdout-marker", stdout[0]["output"])
	require.Equal(t, "", stdout[1]["output"], "the trailing newline's empty record is preserved")

	require.Len(t, stderr, 2)
	require.Equal(t, "stderr-marker", stderr[0]["output"])
}

// TestLogCommandOutputSplitting pins the line splitting directly, including the
// two edge cases the loop has always had.
func TestLogCommandOutputSplitting(t *testing.T) {
	tests := map[string]struct {
		output []byte
		want   []string
	}{
		"empty output logs nothing":  {output: nil, want: nil},
		"a single line":              {output: []byte("one"), want: []string{"one"}},
		"a trailing newline":         {output: []byte("one\n"), want: []string{"one", ""}},
		"several lines":              {output: []byte("one\ntwo\nthree"), want: []string{"one", "two", "three"}},
		"a lone newline":             {output: []byte("\n"), want: []string{"", ""}},
		"interior blank lines abide": {output: []byte("one\n\ntwo"), want: []string{"one", "", "two"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, records := captureRecords(slog.LevelDebug)

			logCommandOutput(logger, tc.output, "the-resource", "the-command", streamStdout)

			got := make([]string, 0, len(tc.want))
			for _, record := range records() {
				got = append(got, record["output"].(string))
			}
			require.Equal(t, tc.want, nilIfEmpty(got))
		})
	}
}

func nilIfEmpty(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}

// TestCommandOutputIsSilentWithoutALogger keeps the busiest site in the tree
// silent when no logger was supplied.
func TestCommandOutputIsSilentWithoutALogger(t *testing.T) {
	require.NotPanics(t, func() {
		logCommandOutput(nil, []byte("output nobody asked to see"), "the-resource", "the-command", streamStdout)

		// A System built without options, and a DefCommand built from one, are
		// both routine and neither may panic.
		sys := New("")
		logCommandOutput(sys.loggerOrDiscard(), []byte("more output"), "the-resource", "the-command", streamStderr)
	})
}
