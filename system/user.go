package system

import (
	"os/user"
	"sort"

	"github.com/aelsabbahy/goss/util/group"
)

type User interface {
	Username() string
	Exists() (interface{}, error)
	UID() (interface{}, error)
	GID() (interface{}, error)
	Groups() ([]string, error)
	Home() (interface{}, error)
}

type DefUser struct {
	username string
}

func NewDefUser(username string, system *System) User {
	return &DefUser{username: username}
}

func (u *DefUser) Username() string {
	return u.username
}

func (u *DefUser) Exists() (interface{}, error) {
	_, err := user.Lookup(u.username)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *DefUser) UID() (interface{}, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return "", nil
	}

	return user.Uid, nil
}

func (u *DefUser) GID() (interface{}, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return "", nil
	}

	return user.Gid, nil
}

func (u *DefUser) Home() (interface{}, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return "", nil
	}

	return user.HomeDir, nil
}

func (u *DefUser) Groups() ([]string, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return nil, err
	}
	groupList, err := group.GetGroupList(user)
	if err != nil {
		return nil, err
	}

	sort.Strings(groupList)
	return groupList, nil
}
