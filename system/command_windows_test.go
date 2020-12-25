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

	if c.Cmd.SysProcAttr.CmdLine != "/c echo hello world" {
		t.Errorf("Command not wrapped properly for Windows cmd.exe. got %s, want: %s", c.Cmd.SysProcAttr.CmdLine, "/c echo hello world")
	}

	if len(c.Cmd.Args) != 1 {
		t.Errorf("Args length should be blank. got: %d, want: %d", len(c.Cmd.Args), 1)
	}
}
