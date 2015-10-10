package system

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/aelsabbahy/goss/util/group"
)

type File interface {
	Path() string
	Exists() (interface{}, error)
	Contains() (io.Reader, error)
	Mode() (interface{}, error)
	Filetype() (interface{}, error)
	Owner() (interface{}, error)
	Group() (interface{}, error)
	LinkedTo() (interface{}, error)
}

type DefFile struct {
	path string
}

func NewDefFile(path string, system *System) File {
	absPath, _ := filepath.Abs(path)
	return &DefFile{path: absPath}
}

func (f *DefFile) Path() string {
	return f.path
}

func (f *DefFile) Exists() (interface{}, error) {
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

func (f *DefFile) Mode() (interface{}, error) {
	fi, err := os.Lstat(f.path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%#o", fi.Mode().Perm()), nil
}

func (f *DefFile) Filetype() (interface{}, error) {
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

func (f *DefFile) Owner() (interface{}, error) {
	fi, err := os.Lstat(f.path)
	if err != nil {
		return "", err
	}

	uid := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Uid)
	user, err := user.LookupId(uid)
	if err != nil {
		return "", err
	}

	return user.Username, nil
}

func (f *DefFile) Group() (interface{}, error) {
	fi, err := os.Lstat(f.path)
	if err != nil {
		return "", err
	}

	gid := fmt.Sprint(fi.Sys().(*syscall.Stat_t).Gid)
	group, err := group.LookupGroupID(gid)
	if err != nil {
		return "", err
	}

	return group.Name, nil
}

func (f *DefFile) LinkedTo() (interface{}, error) {
	dst, err := os.Readlink(f.path)
	if err != nil {
		return "", err
	}
	return dst, nil
}
