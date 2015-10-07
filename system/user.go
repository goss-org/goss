package system

import (
	"os/user"
	"sort"

	"github.com/aelsabbahy/goss/util/group"
)

type User struct {
	username string
	exists   bool
	uid      string
	groups   []string
	home     string
}

func NewUser(username string, system *System) User {
	return User{username: username}
}

func (u *User) Username() string {
	return u.username
}

func (u *User) Exists() (interface{}, error) {
	_, err := user.Lookup(u.username)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *User) UID() (interface{}, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return "", nil
	}

	return user.Uid, nil
}

func (u *User) Gid() (interface{}, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return "", nil
	}

	return user.Gid, nil
}

func (u *User) Home() (interface{}, error) {
	user, err := user.Lookup(u.username)
	if err != nil {
		return "", nil
	}

	return user.HomeDir, nil
}

func (u *User) Groups() ([]string, error) {
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
