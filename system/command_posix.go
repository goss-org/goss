// +build linux darwin !windows

package system

import "github.com/aelsabbahy/goss/util"

const linuxShell string = "sh"

func commandWrapper(cmd string) *util.Command {
	return util.NewCommand(linuxShell, "-c", cmd)
}
