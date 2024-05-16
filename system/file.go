package system

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goss-org/goss/util"
)

type File interface {
	Path() string
	Exists() (bool, error)
	Contents() (io.Reader, error)
	Mode() (string, error)
	Size() (int, error)
	Filetype() (string, error)
	Owner() (string, error)
	Uid() (int, error)
	Group() (string, error)
	Gid() (int, error)
	LinkedTo() (string, error)
	Md5() (string, error)
	Sha256() (string, error)
	Sha512() (string, error)
}

type hashFuncType string

const (
	md5Hash    hashFuncType = "md5"
	sha256Hash              = "sha256"
	sha512Hash              = "sha512"
)

type DefFile struct {
	path     string
	realPath string
	fi       os.FileInfo
	loaded   bool
	err      error
}

func NewDefFile(_ context.Context, path string, system *System, config util.Config) File {
	var err error
	if !strings.HasPrefix(path, "~") {
		path, err = filepath.Abs(path)
	}
	return &DefFile{path: path, err: err}
}

func (f *DefFile) setup() error {
	if f.loaded || f.err != nil {
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

	_, err := os.Lstat(f.realPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (f *DefFile) Contents() (io.Reader, error) {
	if err := f.setup(); err != nil {
		return nil, err
	}

	fh, err := os.Open(f.realPath)
	if err != nil {
		return nil, err
	}
	return fh, nil
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
	case fi.Mode()&os.ModeDevice == os.ModeDevice:
		if fi.Mode()&os.ModeCharDevice == os.ModeCharDevice {
			return "character-device", nil
		}
		return "block-device", nil
	case fi.Mode()&os.ModeNamedPipe == os.ModeNamedPipe:
		return "pipe", nil
	case fi.Mode()&os.ModeSocket == os.ModeSocket:
		return "socket", nil
	case fi.IsDir():
		return "directory", nil
	case fi.Mode().IsRegular():
		return "file", nil
	}
	// FIXME: file as a catchall?
	return "file", nil
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

	var usr *user.User
	var err error
	if f == "~" {
		usr, err = user.Current()
	} else {
		usr, err = user.Lookup(f[1:len(f)])
	}
	if err != nil {
		return "", err
	}
	pathS[0] = usr.HomeDir

	realPath := strings.Join(pathS, "/")
	realPath, err = filepath.Abs(realPath)

	return realPath, err
}

func (f *DefFile) hash(hashFunc hashFuncType) (string, error) {

	if err := f.setup(); err != nil {
		return "", err
	}

	fh, err := os.Open(f.realPath)
	if err != nil {
		return "", err
	}
	defer fh.Close()

	var hash hash.Hash

	switch hashFunc {
	case md5Hash:
		hash = md5.New()
	case sha256Hash:
		hash = sha256.New()
	case sha512Hash:
		hash = sha512.New()
	default:
		return "", fmt.Errorf("Unsupported hash function %s", hashFunc)
	}

	if _, err := io.Copy(hash, fh); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (f *DefFile) Md5() (string, error) {
	return f.hash(md5Hash)
}

func (f *DefFile) Sha256() (string, error) {
	return f.hash(sha256Hash)
}

func (f *DefFile) Sha512() (string, error) {
	return f.hash(sha512Hash)
}

func getUserForUid(uid int) (string, error) {
	if user, err := user.LookupId(strconv.Itoa(uid)); err == nil {
		return user.Username, nil
	}

	cmd := util.NewCommand("getent", "passwd", strconv.Itoa(uid))
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error: no matching entries in passwd file. getent passwd: %v", err)
	}
	userS := strings.Split(cmd.Stdout.String(), ":")[0]

	return userS, nil
}

func getGroupForGid(gid int) (string, error) {
	if group, err := user.LookupGroupId(strconv.Itoa(gid)); err == nil {
		return group.Name, nil
	}

	cmd := util.NewCommand("getent", "group", strconv.Itoa(gid))
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Error: no matching entries in passwd file. getent group: %v", err)
	}
	groupS := strings.Split(cmd.Stdout.String(), ":")[0]

	return groupS, nil
}
