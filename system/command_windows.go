//go:build windows
// +build windows

package system

import "github.com/goss-org/goss/util"

const windowsShell string = "cmd"

func commandWrapper(cmd string) *util.Command {
	return util.NewCommandForWindowsCmd(windowsShell, "/c", cmd)
}
