package system

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"crypto/md5"

	"github.com/aelsabbahy/goss/util"
	"github.com/opencontainers/runc/libcontainer/user"
)

type File interface {
	Path() string
	Exists() (bool, error)
	Contains() (io.Reader, error)
	Mode() (string, error)
	Size() (int, error)
	Filetype() (string, error)
	Owner() (string, error)
	Group() (string, error)
	LinkedTo() (string, error)
	Md5() (string, error)
}

type DefFile struct {
	path     string
	realPath string
	fi       os.FileInfo
	loaded   bool
	err      error
}

func NewDefFile(path string, system *System, config util.Config) File {
	if !strings.HasPrefix(path, "~") {
		// FIXME: we probably shouldn't ignore errors here
		path, _ = filepath.Abs(path)
	}
	return &DefFile{path: path}
}

func (f *DefFile) setup() error {
	if f.loaded {
		return f.err
	}
	f.loaded = true
	if f.realPath, f.err = realPath(f.path); f.err != nil {
		return f.err
	}

	return f.err
}

func (f *DefFile) Path() string {
	return f.path
}

func (f *DefFile) Exists() (bool, error) {
	if err := f.setup(); err != nil {
		return false, err
	}

	if _, err := os.Stat(f.realPath); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func (f *DefFile) Contains() (io.Reader, error) {
	if err := f.setup(); err != nil {
		return nil, err
	}

	fh, err := os.Open(f.realPath)
	if err != nil {
		return nil, err
	}
	return fh, nil
}

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

func (f *DefFile) Size() (int, error) {
	if err := f.setup(); err != nil {
		return 0, err
	}

	fi, err := os.Lstat(f.realPath)
	if err != nil {
		return 0, err
	}

	size := fi.Size()
	return int(size), nil
}

func (f *DefFile) Filetype() (string, error) {
	if err := f.setup(); err != nil {
		return "", err
	}

	fi, err := os.Lstat(f.realPath)
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
	user, err := user.LookupUid(uid)
	if err != nil {
		return "", err
	}

	return user.Name, nil
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
	group, err := user.LookupGid(gid)
	if err != nil {
		return "", err
	}

	return group.Name, nil
}

func (f *DefFile) LinkedTo() (string, error) {
	if err := f.setup(); err != nil {
		return "", err
	}

	dst, err := os.Readlink(f.realPath)
	if err != nil {
		return "", err
	}
	return dst, nil
}

func realPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	pathS := strings.Split(path, "/")
	f := pathS[0]

	var usr user.User
	var err error
	if f == "~" {
		usr, err = user.CurrentUser()
	} else {
		usr, err = user.LookupUser(f[1:len(f)])
	}
	if err != nil {
		return "", err
	}
	pathS[0] = usr.Home

	realPath := strings.Join(pathS, "/")
	realPath, err = filepath.Abs(realPath)

	return realPath, err
}

func (f *DefFile) Md5() (string, error) {
  var result []byte
	if err := f.setup(); err != nil {
		return "", err
	}

	fh, err := os.Open(f.realPath)
	if err != nil {
		return "", err
	}
  defer fh.Close()

	hash := md5.New()
  if _, err := io.Copy(hash, fh); err != nil {
    return string(result), err
  }

  return fmt.Sprintf("%x", hash.Sum(result)), nil
}
