package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
