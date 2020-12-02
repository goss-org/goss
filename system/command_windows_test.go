// +build windows

package system

import (
	"os/exec"
	"testing"
)

func TestCommandWrapper(t *testing.T) {
	t.Parallel()

	c := commandWrapper("echo hello world")
	cmdPath, _ := exec.LookPath(windowsShell)
	if c.Cmd.Path != cmdPath {
		t.Errorf("Command not wrapped properly for Windows os. got %s, want: %s", c.Cmd.Path, cmdPath)
	}
}
