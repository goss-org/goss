package system

import (
	"context"
	"fmt"
	"os/user"
	"sort"
)

func (u *DefUser) Groups(ctx context.Context) ([]string, error) {
	usr, err := user.Lookup(u.username)
	if err != nil {
		return nil, err
	}

	var groupList []string
	ids, err := usr.GroupIds()
	if err != nil {
		return nil, err
	}

	for _, gid := range ids {
		group, err := user.LookupGroupId(gid)
		if err != nil {
			return nil, fmt.Errorf("Unable to find groups for user %v: %v", usr.Username, err)
		}
		groupList = append(groupList, group.Name)
	}

	sort.Strings(groupList)
	return groupList, nil
}
