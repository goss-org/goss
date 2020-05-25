// +build linux darwin !windows

package system

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

func (f *DefFile) Mode() (string, error) {
	fi, err := f.getFileInfo()
	if err != nil {
		return "", err
	}

	stat := fi.Sys().(*syscall.Stat_t)
	mode := fmt.Sprintf("%04o", (stat.Mode & 07777))
	return mode, nil
}

func (f *DefFile) Owner() (string, error) {
	fi, err := f.getFileInfo()
	if err != nil {
		return "", err
	}

	uidS := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Uid)
	uid, err := strconv.Atoi(uidS)
	if err != nil {
		return "", err
	}
	return getUserForUid(uid)
}

func (f *DefFile) Group() (string, error) {
	fi, err := f.getFileInfo()
	if err != nil {
		return "", err
	}

	gidS := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Gid)
	gid, err := strconv.Atoi(gidS)
	if err != nil {
		return "", err
	}
	return getGroupForGid(gid)
}

func (f *DefFile) getFileInfo() (os.FileInfo, error) {
	if err := f.setup(); err != nil {
		return nil, err
	}

	fi, err := os.Lstat(f.realPath)
	if err != nil {
		return nil, err
	}
	return fi, nil
}
