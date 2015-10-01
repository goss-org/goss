package system

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"syscall"

	"github.com/aelsabbahy/goss/util/group"
)

type File struct {
	ID, path, mode, owner, group, content string
	fi                                    os.FileInfo
}

func NewFile(path string, system *System) *File {
	return &File{ID: path, path: path}
}

func (f *File) Path() string {
	return f.path
}

func (f *File) Exists() (interface{}, error) {
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func (f *File) Contains() (io.Reader, error) {
	fh, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	return fh, nil
}

func (f *File) Mode() (interface{}, error) {
	fi, err := os.Lstat(f.path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%#o", fi.Mode().Perm()), nil
}

func (f *File) Filetype() (interface{}, error) {
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

func (f *File) Owner() (interface{}, error) {
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

func (f *File) Group() (interface{}, error) {
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

func (f *File) LinkedTo() (interface{}, error) {
	dst, err := os.Readlink(f.path)
	if err != nil {
		return "", err
	}
	return dst, nil
}
