package system

import "github.com/aelsabbahy/goss/util/group"

type Group struct {
	groupname string
	exists    bool
	gid       string
}

func NewGroup(groupname string, system *System) *Group {
	return &Group{groupname: groupname}
}

func (u *Group) Groupname() string {
	return u.groupname
}

func (u *Group) Exists() (interface{}, error) {
	_, err := group.LookupGroup(u.groupname)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (u *Group) Gid() (interface{}, error) {
	group, err := group.LookupGroup(u.groupname)
	if err != nil {
		return "", nil
	}

	return group.Gid, nil
}
