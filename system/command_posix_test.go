//go:build linux || darwin || !windows
// +build linux darwin !windows

package system

import (
	"context"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
)

func TestCommandWrapper(t *testing.T) {
	t.Parallel()

	c := commandWrapper("echo hello world")
	cmdPath, _ := exec.LookPath(linuxShell)
	if c.Cmd.Path != cmdPath {
		t.Errorf("Command not wrapped properly for OS. got %s, want: %s", c.Cmd.Path, cmdPath)
	}
}

func TestDefCommandEnv(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		command any
		env     map[string]string
		want    string
	}{
		"shell style":      {command: `printf %s "$GOSS_TEST_A"`, env: map[string]string{"GOSS_TEST_A": "one"}, want: "one"},
		"exec style":       {command: []string{"sh", "-c", `printf %s "$GOSS_TEST_B"`}, env: map[string]string{"GOSS_TEST_B": "two"}, want: "two"},
		"overrides parent": {command: `printf %s "$HOME"`, env: map[string]string{"HOME": "/goss-test-home"}, want: "/goss-test-home"},
		"keeps parent env": {command: `printf %s "${PATH:+set}"`, env: map[string]string{"GOSS_TEST_C": "x"}, want: "set"},
		"no env":           {command: `printf %s "${GOSS_TEST_UNSET:-unset}"`, want: "unset"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cmd := NewDefCommand(context.Background(), tc.command, nil, util.Config{Timeout: 5 * time.Second, CommandEnv: tc.env})
			out, err := cmd.Stdout()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, _ := io.ReadAll(out)
			if string(got) != tc.want {
				t.Errorf("stdout = %q, want %q", got, tc.want)
			}
		})
	}
}
