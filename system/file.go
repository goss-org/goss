package system

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/aelsabbahy/goss/util"
	"github.com/opencontainers/runc/libcontainer/user"
)

type File interface {
	Path() string
	Exists() (bool, error)
	Contains() (io.Reader, error)
	Mode() (string, error)
	Filetype() (string, error)
	Owner() (string, error)
	Group() (string, error)
	LinkedTo() (string, error)
}

type DefFile struct {
	path string
}

func NewDefFile(path string, system *System, config util.Config) File {
	absPath, _ := filepath.Abs(path)
	return &DefFile{path: absPath}
}

func (f *DefFile) Path() string {
	return f.path
}

func (f *DefFile) Exists() (bool, error) {
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func (f *DefFile) Contains() (io.Reader, error) {
	fh, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	return fh, nil
}

func (f *DefFile) Mode() (string, error) {
	fi, err := os.Lstat(f.path)
	if err != nil {
		return "", err
	}

	sys := fi.Sys()
	stat := sys.(*syscall.Stat_t)
	mode := fmt.Sprintf("%04o", (stat.Mode & 07777))
	return mode, nil
}

func (f *DefFile) Filetype() (string, error) {
	fi, err := os.Lstat(f.path)
	if err != nil {
		return "", err
	}

	switch {
	case fi.Mode()&os.ModeSymlink == os.ModeSymlink:
		return "symlink", nil
	case fi.IsDir():
		return "directory", nil
	case fi.Mode().IsRegular():
		return "file", nil
	}
	// FIXME: file as a catchall?
	return "file", nil
}

func (f *DefFile) Owner() (string, error) {
	fi, err := os.Lstat(f.path)
	if err != nil {
		return "", err
	}

	uidS := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Uid)
	uid, err := strconv.Atoi(uidS)
	if err != nil {
		return "", err
	}
	user, err := user.LookupUid(uid)
	if err != nil {
		return "", err
	}

	return user.Name, nil
}

func (f *DefFile) Group() (string, error) {
	fi, err := os.Lstat(f.path)
	if err != nil {
		return "", err
	}

	gidS := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Gid)
	gid, err := strconv.Atoi(gidS)
	if err != nil {
		return "", err
	}
	group, err := user.LookupGid(gid)
	if err != nil {
		return "", err
	}

	return group.Name, nil
}

func (f *DefFile) LinkedTo() (string, error) {
	dst, err := os.Readlink(f.path)
	if err != nil {
		return "", err
	}
	return dst, nil
}
