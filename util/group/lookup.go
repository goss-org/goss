package group

// CurrentGroup returns the current group.
func CurrentGroup() (*Group, error) {
	return currentGroup()
}

// LookupGroup looks up a group by name. If the group cannot be found, the
// returned error is of type UnknownGroupError.
func LookupGroup(groupname string) (*Group, error) {
	return lookupGroup(groupname)
}

// LookupGroupID looks up a group by groupid. If the group cannot be found, the
// returned error is of type UnknownGroupIDError.
func LookupGroupID(gid string) (*Group, error) {
	return lookupGroupID(gid)
}

// // In indicates whether the user is a member of the given group.
// func (u *User) In(g *Group) (bool, error) {
// 	return userInGroup(u, g)
// }

// Members returns the list of members of the group.
func (g *Group) Members() ([]string, error) {
	return groupMembers(g)
}

//func GetGroupList(groupname string) (*Group, error) {
//	return getGroupList(groupname)
//}
