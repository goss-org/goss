package system

import (
	"context"
	"os/user"
	"strconv"

	"github.com/goss-org/goss/util"
)

type User interface {
	Username() string
	Exists(context.Context) (bool, error)
	UID(context.Context) (int, error)
	GID(context.Context) (int, error)
	Groups(context.Context) ([]string, error)
	Home(context.Context) (string, error)
	Shell(context.Context) (string, error)
}

type DefUser struct {
	username string
}

func NewDefUser(username string, system *System, config util.Config) User {
	return &DefUser{username: username}
}

func (u *DefUser) Username() string {
	return u.username
}

func (u *DefUser) Exists(ctx context.Context) (bool, error) {
	_, err := user.Lookup(u.username)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *DefUser) UID(ctx context.Context) (int, error) {
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

func (u *DefUser) GID(ctx context.Context) (int, error) {
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

func (u *DefUser) Home(ctx context.Context) (string, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return "", err
	}

	return user.HomeDir, nil
}
