package resource

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
	"gopkg.in/yaml.v3"
)

type fakePackage struct {
	name        string
	installedFn func() (bool, error)
	versionsFn  func() ([]string, error)
}

func (f *fakePackage) Name() string { return f.name }

func (f *fakePackage) Exists() (bool, error) {
	return f.Installed()
}

func (f *fakePackage) Installed() (bool, error) {
	return f.installedFn()
}

func (f *fakePackage) Versions() ([]string, error) {
	return f.versionsFn()
}

type fakeDNS struct {
	host         string
	server       string
	qtype        string
	resolvableFn func() (bool, error)
	addrsFn      func() ([]string, error)
}

func (f *fakeDNS) Host() string              { return f.host }
func (f *fakeDNS) Server() string            { return f.server }
func (f *fakeDNS) Qtype() string             { return f.qtype }
func (f *fakeDNS) Exists() (bool, error)     { return false, nil }
func (f *fakeDNS) Resolvable() (bool, error) { return f.resolvableFn() }
func (f *fakeDNS) Addrs() ([]string, error)  { return f.addrsFn() }

type fakeCommand struct {
	command      string
	exitStatusFn func() (int, error)
	stdoutFn     func() (io.Reader, error)
	stderrFn     func() (io.Reader, error)
}

func (f *fakeCommand) Command() string            { return f.command }
func (f *fakeCommand) Exists() (bool, error)      { return true, nil }
func (f *fakeCommand) ExitStatus() (int, error)   { return f.exitStatusFn() }
func (f *fakeCommand) Stdout() (io.Reader, error) { return f.stdoutFn() }
func (f *fakeCommand) Stderr() (io.Reader, error) { return f.stderrFn() }

func TestRetryDelayUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Duration
	}{
		{
			name:  "integers are interpreted as milliseconds",
			input: "retry_count: 1\nretry_delay: 5\ninstalled: true\n",
			want:  5 * time.Millisecond,
		},
		{
			name:  "floats are treated as fractional milliseconds",
			input: "retry_count: 1\nretry_delay: 0.5\ninstalled: true\n",
			want:  500 * time.Microsecond,
		},
		{
			name:  "duration strings support millisecond granularity",
			input: "retry_count: 1\nretry_delay: 250ms\ninstalled: true\n",
			want:  250 * time.Millisecond,
		},
		{
			name:  "duration strings support seconds",
			input: "retry_count: 1\nretry_delay: 2s\ninstalled: true\n",
			want:  2 * time.Second,
		},
		{
			name:  "duration strings support minutes",
			input: "retry_count: 1\nretry_delay: 1m\ninstalled: true\n",
			want:  time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pkg Package
			if err := yaml.Unmarshal([]byte(tt.input), &pkg); err != nil {
				t.Fatalf("yaml unmarshal failed: %v", err)
			}

			if got := pkg.RetryDelay.Duration(); got != tt.want {
				t.Fatalf("RetryDelay = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRetryDelayUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Duration
	}{
		{
			name:  "numeric json values are milliseconds",
			input: `{"retry_count":1,"retry_delay":75,"exit-status":0}`,
			want:  75 * time.Millisecond,
		},
		{
			name:  "duration strings still work in json",
			input: `{"retry_count":1,"retry_delay":"75ms","exit-status":0}`,
			want:  75 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd Command
			if err := json.Unmarshal([]byte(tt.input), &cmd); err != nil {
				t.Fatalf("json unmarshal failed: %v", err)
			}

			if got := cmd.RetryDelay.Duration(); got != tt.want {
				t.Fatalf("RetryDelay = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRetryDelayMarshalYAML(t *testing.T) {
	pkg := Package{
		Installed:  true,
		RetryCount: 1,
		RetryDelay: RetryDelay(500 * time.Millisecond),
	}

	out, err := yaml.Marshal(pkg)
	if err != nil {
		t.Fatalf("yaml marshal failed: %v", err)
	}

	if !strings.Contains(string(out), "retry_delay: 500") {
		t.Fatalf("expected retry_delay to be rendered as millisecond integer, got:\n%s", string(out))
	}
}

func TestValidateValueWithRetrySupportsSubsecondDelay(t *testing.T) {
	start := time.Now()
	attempts := 0

	result := ValidateValueWithRetry(&FakeResource{id: "retry"}, "installed", true, func() (any, error) {
		attempts++
		return attempts >= 2, nil
	}, false, 1, RetryDelay(25*time.Millisecond))

	if result.Result != SUCCESS {
		t.Fatalf("expected retry to succeed, got %d", result.Result)
	}

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}

	if elapsed := time.Since(start); elapsed >= 500*time.Millisecond {
		t.Fatalf("expected sub-second retry delay, elapsed %s", elapsed)
	}
}

func TestPackageValidateRetries(t *testing.T) {
	installedCalls := 0
	versionCalls := 0
	sys := &system.System{
		NewPackage: func(_ context.Context, name string, _ *system.System, _ util.Config) system.Package {
			return &fakePackage{
				name: name,
				installedFn: func() (bool, error) {
					installedCalls++
					return installedCalls >= 2, nil
				},
				versionsFn: func() ([]string, error) {
					versionCalls++
					return []string{"1.0.0"}, nil
				},
			}
		},
	}

	pkg := &Package{
		id:         "nginx",
		Installed:  true,
		Versions:   []interface{}{"1.0.0"},
		RetryCount: 1,
		RetryDelay: RetryDelay(10 * time.Millisecond),
	}

	results := pkg.Validate(sys)

	if installedCalls != 2 {
		t.Fatalf("expected installed to be retried once, got %d calls", installedCalls)
	}

	if versionCalls != 1 {
		t.Fatalf("expected versions to run once after installed succeeds, got %d calls", versionCalls)
	}

	if !allTestsPassed(results) {
		t.Fatalf("expected package retry validation to pass, got %+v", results)
	}
}

func TestDNSValidateRetries(t *testing.T) {
	resolvableCalls := 0
	addrCalls := 0
	sys := &system.System{
		NewDNS: func(_ context.Context, host string, _ *system.System, _ util.Config) system.DNS {
			return &fakeDNS{
				host: host,
				resolvableFn: func() (bool, error) {
					resolvableCalls++
					return resolvableCalls >= 2, nil
				},
				addrsFn: func() ([]string, error) {
					addrCalls++
					return []string{"127.0.0.1"}, nil
				},
			}
		},
	}

	dns := &DNS{
		id:         "localhost",
		Resolvable: true,
		Addrs:      []interface{}{"127.0.0.1"},
		RetryCount: 1,
		RetryDelay: RetryDelay(10 * time.Millisecond),
	}

	results := dns.Validate(sys)

	if resolvableCalls != 2 {
		t.Fatalf("expected resolvable to be retried once, got %d calls", resolvableCalls)
	}

	if addrCalls != 1 {
		t.Fatalf("expected addrs to run once after resolvable succeeds, got %d calls", addrCalls)
	}

	if !allTestsPassed(results) {
		t.Fatalf("expected dns retry validation to pass, got %+v", results)
	}
}

func TestCommandValidateRetries(t *testing.T) {
	commandCalls := 0
	sys := &system.System{
		NewCommand: func(_ context.Context, command string, _ *system.System, _ util.Config) system.Command {
			commandCalls++
			attempt := commandCalls
			return &fakeCommand{
				command: command,
				exitStatusFn: func() (int, error) {
					if attempt == 1 {
						return 1, nil
					}
					return 0, nil
				},
				stdoutFn: func() (io.Reader, error) {
					return bytes.NewBufferString("ok\n"), nil
				},
				stderrFn: func() (io.Reader, error) {
					return bytes.NewBuffer(nil), nil
				},
			}
		},
	}

	cmd := &Command{
		id:         "echo ok",
		Exec:       "echo ok",
		ExitStatus: 0,
		Stdout:     []interface{}{"ok"},
		RetryCount: 1,
		RetryDelay: RetryDelay(10 * time.Millisecond),
	}

	results := cmd.Validate(sys)

	if commandCalls != 2 {
		t.Fatalf("expected command to be retried once, got %d attempts", commandCalls)
	}

	if !allTestsPassed(results) {
		t.Fatalf("expected command retry validation to pass, got %+v", results)
	}
}
