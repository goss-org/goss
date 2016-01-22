package system

import (
	"strconv"

	"github.com/aelsabbahy/goss/util"
	"github.com/opencontainers/runc/libcontainer/user"
)

type Group interface {
	Groupname() string
	Exists() (interface{}, error)
	Gid() (interface{}, error)
}

type DefGroup struct {
	groupname string
}

func NewDefGroup(groupname string, system *System, config util.Config) Group {
	return &DefGroup{groupname: groupname}
}

func (u *DefGroup) Groupname() string {
	return u.groupname
}

func (u *DefGroup) Exists() (interface{}, error) {
	_, err := user.LookupGroup(u.groupname)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *DefGroup) Gid() (interface{}, error) {
	group, err := user.LookupGroup(u.groupname)
	if err != nil {
		return "", nil
	}

	return strconv.Itoa(group.Gid), nil
}
