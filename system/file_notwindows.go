// +build !windows

package system

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/aelsabbahy/goss/util"
	"github.com/opencontainers/runc/libcontainer/user"
)

func (f *DefFile) Mode() (string, error) {
	if err := f.setup(); err != nil {
		return "", err
	}

	fi, err := os.Lstat(f.realPath)
	if err != nil {
		return "", err
	}

	sys := fi.Sys()
	stat := sys.(*syscall.Stat_t)
	mode := fmt.Sprintf("%04o", (stat.Mode & 07777))
	return mode, nil
}

func (f *DefFile) Owner() (string, error) {
	if err := f.setup(); err != nil {
		return "", err
	}

	fi, err := os.Lstat(f.realPath)
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
	if err := f.setup(); err != nil {
		return "", err
	}

	fi, err := os.Lstat(f.realPath)
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
