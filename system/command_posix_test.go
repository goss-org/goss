// +build linux darwin !windows

package system

import (
	"os/exec"
	"testing"
)

func TestCommandWrapper(t *testing.T) {
	t.Parallel()

	c := commandWrapper("echo hello world")
	cmdPath, _ := exec.LookPath(linuxShell)
	if c.Cmd.Path != cmdPath {
		t.Errorf("Command not wrapped properly for OS. got %s, want: %s", c.Cmd.Path, cmdPath)
	}
}
