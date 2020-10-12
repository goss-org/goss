// +build windows

package system

import "github.com/aelsabbahy/goss/util"

func commandWrapper(cmd string) *util.Command {
	return util.NewCommand("cmd", "/c", cmd)
}
