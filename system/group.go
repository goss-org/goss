package system

import (
	"context"
	"os/user"
	"strconv"

	"github.com/goss-org/goss/util"
)

type Group interface {
	Groupname() string
	Exists(context.Context) (bool, error)
	GID(context.Context) (int, error)
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

func (u *DefGroup) Exists(ctx context.Context) (bool, error) {
	_, err := user.LookupGroup(u.groupname)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *DefGroup) GID(ctx context.Context) (int, error) {
	group, err := user.LookupGroup(u.groupname)
	if err != nil {
		return 0, err
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return 0, err
	}

	return gid, nil
}
