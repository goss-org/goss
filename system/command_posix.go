//go:build linux || darwin || !windows
// +build linux darwin !windows

package system

import "github.com/goss-org/goss/util"

const linuxShell string = "sh"

func commandWrapper(cmd string) *util.Command {
	return util.NewCommand(linuxShell, "-c", cmd)
}
