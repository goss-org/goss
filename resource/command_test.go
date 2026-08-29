package resource

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/goss-org/goss/util"
	"go.yaml.in/yaml/v3"
)

func newFakeCommand(stdout, stderr string) *fakeCommand {
	return &fakeCommand{
		command:      util.ExecCommand{CmdStr: "some-command"},
		exitStatusFn: func() (int, error) { return 0, nil },
		stdoutFn:     func() (io.Reader, error) { return strings.NewReader(stdout), nil },
		stderrFn:     func() (io.Reader, error) { return strings.NewReader(stderr), nil },
	}
}

func TestNewCommandDefaultCapturesTrimmedLines(t *testing.T) {
	sysCommand := newFakeCommand("line one\n  indented two\nline three\n", "")

	c, err := NewCommand(sysCommand, util.Config{})
	if err != nil {
		t.Fatalf("NewCommand returned error: %v", err)
	}

	want := []string{"line one", "indented two", "line three"}
	if got, ok := c.Stdout.([]string); !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("Stdout = %#v, want %#v", c.Stdout, want)
	}
}

func TestNewCommandExactMatchCapturesRawString(t *testing.T) {
	raw := "line one\n  indented two\nline three\n"
	sysCommand := newFakeCommand(raw, "")

	c, err := NewCommand(sysCommand, util.Config{ExactMatch: true})
	if err != nil {
		t.Fatalf("NewCommand returned error: %v", err)
	}

	if got, ok := c.Stdout.(string); !ok || got != raw {
		t.Fatalf("Stdout = %#v, want exact %#v", c.Stdout, raw)
	}
}

func TestNewCommandExactMatchLeavesEmptyOutputUnset(t *testing.T) {
	sysCommand := newFakeCommand("", "")

	c, err := NewCommand(sysCommand, util.Config{ExactMatch: true})
	if err != nil {
		t.Fatalf("NewCommand returned error: %v", err)
	}

	if got, ok := c.Stdout.(string); !ok || got != "" {
		t.Fatalf("Stdout = %#v, want empty string", c.Stdout)
	}
	if got, ok := c.Stderr.(string); !ok || got != "" {
		t.Fatalf("Stderr = %#v, want empty string", c.Stderr)
	}
}

func TestNewCommandExactMatchHonorsIgnoreList(t *testing.T) {
	sysCommand := newFakeCommand("some output\n", "some error\n")

	c, err := NewCommand(sysCommand, util.Config{ExactMatch: true, IgnoreList: []string{"stdout"}})
	if err != nil {
		t.Fatalf("NewCommand returned error: %v", err)
	}

	if got, ok := c.Stdout.(string); !ok || got != "" {
		t.Fatalf("Stdout = %#v, want empty string when ignored", c.Stdout)
	}
	if got, ok := c.Stderr.(string); !ok || got != "some error\n" {
		t.Fatalf("Stderr = %#v, want captured raw string", c.Stderr)
  }
}

func TestCommandGetExec(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  any
	}{
		{
			name:  "no exec falls back to the id",
			input: "foo:\n  exit-status: 0\n",
			want:  "foo",
		},
		{
			name:  "string exec is shell style",
			input: "foo:\n  exit-status: 0\n  exec: echo hi\n",
			want:  "echo hi",
		},
		{
			name:  "list exec is exec style",
			input: "foo:\n  exit-status: 0\n  exec: [/bin/echo, hello world]\n",
			want:  []string{"/bin/echo", "hello world"},
		},
		{
			name:  "bool scalar exec keeps legacy string behavior",
			input: "foo:\n  exit-status: 0\n  exec: true\n",
			want:  "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m CommandMap
			if err := yaml.Unmarshal([]byte(tt.input), &m); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			cmd, ok := m["foo"]
			if !ok {
				t.Fatalf("expected foo resource, got %+v", m)
			}
			got := cmd.GetExec()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetExec() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCommandInvalidExec(t *testing.T) {
	input := "foo:\n  exit-status: 0\n  exec: {a: b}\n"
	var m CommandMap
	if err := yaml.Unmarshal([]byte(input), &m); err == nil {
		t.Fatalf("expected error for mapping exec, got %+v", m)
	}
}
