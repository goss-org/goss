// +build !windows

package system

import (
	"os"

	"golang.org/x/sys/unix"
)

func (u *DefDiskUsage) Exists() (bool, error) {
	if u.err == nil {
		return true, nil
	}
	if os.IsNotExist(u.err) {
		return false, nil
	}
	return false, u.err
}

func (u *DefDiskUsage) Calculate() {
	fd, err := os.Open(u.path)
	if err != nil {
		u.err = err
		return
	}
	var s unix.Statfs_t
	u.err = unix.Fstatfs(int(fd.Fd()), &s)
	u.totalBytes = s.Blocks * uint64(s.Bsize)
	u.freeBytes = s.Bfree * uint64(s.Bsize)
}
