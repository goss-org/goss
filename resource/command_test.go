package resource

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/goss-org/goss/util"
)

func newFakeCommand(stdout, stderr string) *fakeCommand {
	return &fakeCommand{
		command:      "some-command",
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
