package group

import (
	"strconv"
)

var implemented = true // set to false by lookup_stubs.go's init

// Group type
type Group struct {
	Gid  string // group id
	Name string // group name
}

// UnknownGroupIDError is returned by LookupGroupID when
// a group cannot be found.
type UnknownGroupIDError int

func (e UnknownGroupIDError) Error() string {
	return "group: unknown groupid " + strconv.Itoa(int(e))
}

// UnknownGroupError is returned by LookupGroup when
// a group cannot be found.
type UnknownGroupError string

func (e UnknownGroupError) Error() string {
	return "group: unknown group " + string(e)
}
