//go:build linux || darwin || !windows
// +build linux darwin !windows

package system

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

func (f *DefFile) Mode(ctx context.Context) (string, error) {
	mode, err := f.getFileInfo(func(fi os.FileInfo) string {
		stat := fi.Sys().(*syscall.Stat_t)
		return fmt.Sprintf("%04o", (stat.Mode & 07777))
	})
	if err != nil {
		return "", err
	}

	return mode, nil
}

func (f *DefFile) Owner(ctx context.Context) (string, error) {
	uidS, err := f.getFileInfo(func(fi os.FileInfo) string {
		return fmt.Sprint(fi.Sys().(*syscall.Stat_t).Uid)
	})
	if err != nil {
		return "", err
	}

	uid, err := strconv.Atoi(uidS)
	if err != nil {
		return "", err
	}
	return getUserForUid(uid)
}

func (f *DefFile) Group(ctx context.Context) (string, error) {
	gidS, err := f.getFileInfo(func(fi os.FileInfo) string {
		return fmt.Sprint(fi.Sys().(*syscall.Stat_t).Gid)
	})
	if err != nil {
		return "", err
	}

	gid, err := strconv.Atoi(gidS)
	if err != nil {
		return "", err
	}
	return getGroupForGid(gid)
}

func (f *DefFile) getFileInfo(selectorFunc func(os.FileInfo) string) (string, error) {
	if err := f.setup(); err != nil {
		return "", err
	}

	fi, err := os.Lstat(f.realPath)
	if err != nil {
		return "", err
	}
	return selectorFunc(fi), nil
}
