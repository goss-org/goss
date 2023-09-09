package system

import (
	"context"
	"os/user"
	"strconv"

	"github.com/goss-org/goss/util"
)

type User interface {
	Username() string
	Exists() (bool, error)
	UID() (int, error)
	GID() (int, error)
	Groups() ([]string, error)
	Home() (string, error)
	Shell() (string, error)
}

type DefUser struct {
	username string
}

func NewDefUser(_ context.Context, username string, system *System, config util.Config) User {
	return &DefUser{username: username}
}

func (u *DefUser) Username() string {
	return u.username
}

func (u *DefUser) Exists() (bool, error) {
	_, err := user.Lookup(u.username)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *DefUser) UID() (int, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return 0, err
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return 0, err
	}

	return uid, nil
}

func (u *DefUser) GID() (int, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return 0, err
	}

	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return 0, err
	}

	return gid, nil
}

func (u *DefUser) Home() (string, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return "", err
	}

	return user.HomeDir, nil
}
