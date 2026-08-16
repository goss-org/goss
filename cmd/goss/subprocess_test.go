package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// The terminal paths of the CLI end the process, and the platform gate is one of
// them, so the only honest way to assert on them is to run the real binary.

var (
	buildOnce sync.Once
	buildDir  string
	buildPath string
	buildErr  error
)

// gossBinary builds cmd/goss once per test binary run and returns its path.
func gossBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		buildDir, buildErr = os.MkdirTemp("", "goss-cli-test")
		if buildErr != nil {
			return
		}
		buildPath = filepath.Join(buildDir, "goss")
		if runtime.GOOS == "windows" {
			buildPath += ".exe"
		}

		cmd := exec.Command("go", "build", "-o", buildPath, ".")
		if out, err := cmd.CombinedOutput(); err != nil {
			buildErr = &buildFailure{err: err, output: string(out)}
		}
	})

	if buildErr != nil {
		t.Fatalf("building the goss binary: %v", buildErr)
	}
	return buildPath
}

type buildFailure struct {
	err    error
	output string
}

func (b *buildFailure) Error() string {
	return b.err.Error() + ": " + b.output
}

func TestMain(m *testing.M) {
	code := m.Run()
	if buildDir != "" {
		_ = os.RemoveAll(buildDir)
	}
	os.Exit(code)
}

type cliResult struct {
	exitCode int
	stdout   string
	stderr   string
}

// runBinary runs the built goss binary with a deliberately minimal environment,
// so that a developer's own GOSS_* settings cannot change what a test observes.
func runBinary(t *testing.T, env []string, args ...string) cliResult {
	t.Helper()

	cmd := exec.Command(gossBinary(t), args...)
	cmd.Env = append([]string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}, env...)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	code := 0
	var exitErr *exec.ExitError
	switch err := cmd.Run(); {
	case err == nil:
	case errors.As(err, &exitErr):
		code = exitErr.ExitCode()
	default:
		t.Fatalf("running %v: %v", args, err)
	}

	return cliResult{exitCode: code, stdout: stdout.String(), stderr: stderr.String()}
}

// TestAlphaGateRunsBeforeLevelValidation pins precedence when both conditions
// are wrong at once: the platform is unsupported and the level is unusable. On
// the alpha platforms the gate must win, because it runs before any runtime
// configuration is constructed. Everywhere else the level error is what
// surfaces.
func TestAlphaGateRunsBeforeLevelValidation(t *testing.T) {
	got := runBinary(t, nil, "--log-level", "VERBOSE", "serve")

	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		if got.stderr == "" {
			t.Fatalf("expected a diagnostic, got none (exit %d)", got.exitCode)
		}
		if !strings.Contains(got.stderr, "Terminating.") {
			t.Errorf("expected the alpha diagnostic, got %q", got.stderr)
		}
		if strings.Contains(got.stderr, "unsupported log level") {
			t.Errorf("level validation should not have run: %q", got.stderr)
		}
	} else {
		if !strings.Contains(got.stderr, "unsupported log level: VERBOSE") {
			t.Errorf("expected the level error naming the value, got %q", got.stderr)
		}
		if strings.Contains(got.stderr, "Terminating.") {
			t.Errorf("the alpha gate should not fire on %s: %q", runtime.GOOS, got.stderr)
		}
	}

	if got.exitCode != 1 {
		t.Errorf("expected exit status 1, got %d", got.exitCode)
	}
	if got.stdout != "" {
		t.Errorf("nothing should reach stdout, got %q", got.stdout)
	}
}

// TestInvalidLevelFailsPastTheAlphaGate completes the pair: with the platform
// opt-in given, the level error is what stops the run on every host.
func TestInvalidLevelFailsPastTheAlphaGate(t *testing.T) {
	got := runBinary(t, []string{"GOSS_USE_ALPHA=1"}, "--log-level", "VERBOSE", "serve")

	if !strings.Contains(got.stderr, "unsupported log level: VERBOSE") {
		t.Errorf("expected the level error naming the value, got %q", got.stderr)
	}
	if got.exitCode != 1 {
		t.Errorf("expected exit status 1, got %d", got.exitCode)
	}
	if got.stdout != "" {
		t.Errorf("nothing should reach stdout, got %q", got.stdout)
	}
}

// TestTerminalErrorDiagnostic pins the terminal path. When an action returns an
// error, main still reports it through the standard logger and exits 1. Nothing
// about that path changes here, and this pins it because the standard logger no
// longer has setLogLevel's writer underneath it: the envelope is now Go's
// default, local time and all.
func TestTerminalErrorDiagnostic(t *testing.T) {
	// A gossfile that cannot exist, so the failure is the operation's and not a
	// flag parse error, which would print usage instead.
	missing := filepath.Join(t.TempDir(), "definitely-not-here.yaml")

	got := runBinary(t, []string{"GOSS_USE_ALPHA=1"}, "--gossfile", missing, "render")

	if got.exitCode != 1 {
		t.Errorf("expected exit status 1, got %d", got.exitCode)
	}
	if got.stdout != "" {
		t.Errorf("the diagnostic does not belong on stdout, got %q", got.stdout)
	}
	if !strings.Contains(got.stderr, "definitely-not-here.yaml") {
		t.Errorf("expected the diagnostic to name the input, got %q", got.stderr)
	}
	// Exactly one line, and it is the standard logger's, not a structured
	// record: the terminal path has no injected logger to use. Counting matters
	// because the failure that would follow a half-finished conversion is the
	// same error reported twice, once by each logger, which a content check
	// would pass.
	lines := strings.Split(strings.TrimSpace(got.stderr), "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly one diagnostic, got %d: %q", len(lines), got.stderr)
	}
	if strings.Contains(got.stderr, "level=ERROR") {
		t.Errorf("the terminal path should not be emitting slog records: %q", got.stderr)
	}
	// The standard logger's own envelope: a date and time, then the message.
	// Anything else means something reconfigured it on the way out.
	if !regexp.MustCompile(`^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} `).MatchString(lines[0]) {
		t.Errorf("expected the standard logger's envelope, got %q", lines[0])
	}
}

// TestStdoutCarriesNoLogRecords is the behavioural half of the stderr
// guarantee, driven through the real binary so that stdout and stderr are
// genuinely separate streams.
func TestStdoutCarriesNoLogRecords(t *testing.T) {
	spec := filepath.Join(t.TempDir(), "goss.yaml")
	contents := "command:\n  probe:\n    exec: \"echo rendered-marker\"\n    exit-status: 0\n"
	if err := os.WriteFile(spec, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	got := runBinary(t, []string{"GOSS_USE_ALPHA=1"},
		"--log-level", "TRACE", "--gossfile", spec, "render")

	if got.exitCode != 0 {
		t.Fatalf("expected success, got exit %d: %s", got.exitCode, got.stderr)
	}
	if !strings.Contains(got.stdout, "rendered-marker") {
		t.Errorf("the rendered gossfile should reach stdout, got %q", got.stdout)
	}
	for _, marker := range []string{"level=", "msg="} {
		if strings.Contains(got.stdout, marker) {
			t.Errorf("stdout carries a log record: %q", got.stdout)
		}
	}
}

// TestInvalidLevelFromEnvironmentFails covers the same path with the value
// arriving from GOSS_LOGLEVEL, which is indistinguishable from the flag.
func TestInvalidLevelFromEnvironmentFails(t *testing.T) {
	got := runBinary(t, []string{"GOSS_USE_ALPHA=1", "GOSS_LOGLEVEL=VERBOSE"}, "render")

	if !strings.Contains(got.stderr, "unsupported log level: VERBOSE") {
		t.Errorf("expected the level error naming the value, got %q", got.stderr)
	}
	if got.exitCode != 1 {
		t.Errorf("expected exit status 1, got %d", got.exitCode)
	}
}
