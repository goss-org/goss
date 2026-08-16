package goss

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// Serve registers its endpoints on http.DefaultServeMux and has no shutdown
// path, so a second call in the same process panics on a duplicate pattern.
// Driving the real root therefore means driving it in a subprocess, one call
// per process, with a listen address that cannot succeed so the call returns
// instead of blocking.

const (
	serveSubprocessEnv = "GOSS_TEST_SERVE_SUBPROCESS"
	// A port number out of range fails in the listener rather than depending on
	// what else is running on the machine.
	unbindableAddress = "127.0.0.1:99999"
)

// TestServeRootInSubprocess is the child side. It runs only when re-executed by
// its parent below, and asserts nothing itself: the parent reads its output.
func TestServeRootInSubprocess(t *testing.T) {
	mode := os.Getenv(serveSubprocessEnv)
	if mode == "" {
		t.Skip("only runs when re-executed by TestServeRootDrives")
	}

	opts := []util.ConfigOption{
		util.WithSpecFile(filepath.Join("testdata", "matching_basic.yaml")),
		util.WithOutputFormat("json"),
	}
	if mode == "with-logger" {
		// The parent reads this child's stderr, so the records have to leave the
		// process rather than being captured in memory.
		opts = append(opts, util.WithLogger(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:       util.LevelTrace,
			ReplaceAttr: util.ReplaceTraceLevel,
		}))))
	}

	config, err := util.NewConfig(opts...)
	require.NoError(t, err)
	config.ListenAddress = unbindableAddress

	err = Serve(config)
	require.Error(t, err, "an unbindable address should fail rather than block")
}

// TestServeRootDrives is the parent. Both rows run the whole root, which is the
// only way the startup record and the no-logger silence guarantee are covered
// at all.
func TestServeRootDrives(t *testing.T) {
	tests := map[string]struct {
		mode        string
		wantRecords bool
	}{
		"with an injected logger": {mode: "with-logger", wantRecords: true},
		"without a logger":        {mode: "no-logger", wantRecords: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			stdout, stderr := runServeSubprocess(t, tc.mode)

			require.Empty(t, stdout, "nothing goss logs belongs on stdout")

			if !tc.wantRecords {
				require.Empty(t, stderr,
					"with no logger the whole root must be silent, including the startup record")
				return
			}

			listening := recordsWithMessage(decodeJSONRecords(stderr), "server listening")
			require.Len(t, listening, 1, "stderr was: %s", stderr)
			require.Equal(t, "INFO", listening[0]["level"])
			require.Equal(t, unbindableAddress, listening[0]["listen_addr"])
		})
	}
}

func runServeSubprocess(t *testing.T, mode string) (stdout, stderr string) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=TestServeRootInSubprocess", "-test.v=false")
	cmd.Env = append(os.Environ(), serveSubprocessEnv+"="+mode)

	var out, errOut strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	done := make(chan error, 1)
	require.NoError(t, cmd.Start())
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		require.NoError(t, err, "child failed:\nstdout: %s\nstderr: %s", out.String(), errOut.String())
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("the child process did not return; Serve should not have blocked")
	}

	// The child is a test binary, so its own PASS/ok lines are on stdout.
	return stripTestOutput(out.String()), errOut.String()
}

// stripTestOutput removes the go test harness's own lines, leaving whatever the
// code under test wrote.
func stripTestOutput(raw string) string {
	var kept []string
	for line := range strings.SplitSeq(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "",
			trimmed == "PASS",
			strings.HasPrefix(trimmed, "ok "),
			strings.HasPrefix(trimmed, "=== RUN"),
			strings.HasPrefix(trimmed, "--- PASS"),
			strings.HasPrefix(trimmed, "testing:"):
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}
