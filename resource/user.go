package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type User struct {
	Username string   `json:"-"`
	Exists   bool     `json:"exists"`
	UID      string   `json:"uid,omitempty"`
	GID      string   `json:"gid,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Home     string   `json:"home,omitempty"`
}

func (u *User) ID() string      { return u.Username }
func (u *User) SetID(id string) { u.Username = id }

func (u *User) Validate(sys *system.System) []TestResult {
	sysuser := sys.NewUser(u.Username, sys, util.Config{})

	var results []TestResult

	results = append(results, ValidateValue(u, "exists", u.Exists, sysuser.Exists))

	if u.UID != "" {
		results = append(results, ValidateValue(u, "uid", u.UID, sysuser.UID))
	}
	if u.GID != "" {
		results = append(results, ValidateValue(u, "gid", u.GID, sysuser.GID))
	}
	if u.Home != "" {
		results = append(results, ValidateValue(u, "home", u.Home, sysuser.Home))
	}
	if len(u.Groups) > 0 {
		results = append(results, ValidateValues(u, "groups", u.Groups, sysuser.Groups))
	}

	return results
}

func NewUser(sysUser system.User, config util.Config) (*User, error) {
	username := sysUser.Username()
	exists, _ := sysUser.Exists()
	u := &User{
		Username: username,
		Exists:   exists.(bool),
	}
	if !contains(config.IgnoreList, "uid") {
		uid, _ := sysUser.UID()
		u.UID = uid.(string)

	}
	if !contains(config.IgnoreList, "gid") {
		gid, _ := sysUser.GID()
		u.GID = gid.(string)
	}
	if !contains(config.IgnoreList, "groups") {
		groups, _ := sysUser.Groups()
		u.Groups = groups
	}
	if !contains(config.IgnoreList, "home") {
		home, _ := sysUser.Home()
		u.Home = home.(string)
	}
	return u, nil
}
