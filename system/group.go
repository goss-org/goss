package system

import (
	"context"
	"fmt"
	"os/user"
	"strconv"

	"github.com/goss-org/goss/util"
)

type Group interface {
	Groupname() string
	Exists() (bool, error)
	GID() (int, error)
}

type DefGroup struct {
	groupname string
}

func NewDefGroup(_ context.Context, groupname interface{}, system *System, config util.Config) (Group, error) {
	strGroupname, ok := groupname.(string)
	if !ok {
		return nil, fmt.Errorf("groupname must be of type string")
	}
	return newDefGroup(nil, strGroupname, system, config), nil
}

func newDefGroup(_ context.Context, groupname string, system *System, config util.Config) Group {
	return &DefGroup{groupname: groupname}
}

func (u *DefGroup) Groupname() string {
	return u.groupname
}

func (u *DefGroup) Exists() (bool, error) {
	_, err := user.LookupGroup(u.groupname)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *DefGroup) GID() (int, error) {
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
