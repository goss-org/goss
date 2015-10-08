package system

import "github.com/aelsabbahy/goss/util/group"

type Group interface {
	Groupname() string
	Exists() (interface{}, error)
	Gid() (interface{}, error)
}

type DefGroup struct {
	groupname string
	exists    bool
	gid       string
}

func NewDefGroup(groupname string, system *System) Group {
	return &DefGroup{groupname: groupname}
}

func (u *DefGroup) Groupname() string {
	return u.groupname
}

func (u *DefGroup) Exists() (interface{}, error) {
	_, err := group.LookupGroup(u.groupname)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *DefGroup) Gid() (interface{}, error) {
	group, err := group.LookupGroup(u.groupname)
	if err != nil {
		return "", nil
	}

	return group.Gid, nil
}
