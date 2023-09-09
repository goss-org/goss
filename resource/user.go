package resource

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type User struct {
	Title    string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta     meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id       string  `json:"-" yaml:"-"`
	Username string  `json:"username,omitempty" yaml:"username,omitempty"`
	Exists   matcher `json:"exists" yaml:"exists"`
	UID      matcher `json:"uid,omitempty" yaml:"uid,omitempty"`
	GID      matcher `json:"gid,omitempty" yaml:"gid,omitempty"`
	Groups   matcher `json:"groups,omitempty" yaml:"groups,omitempty"`
	Home     matcher `json:"home,omitempty" yaml:"home,omitempty"`
	Shell    matcher `json:"shell,omitempty" yaml:"shell,omitempty"`
	Skip     bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	UserResourceKey  = "user"
	UserResourceName = "User"
)

func init() {
	registerResource(UserResourceKey, &User{})
}

func (u *User) ID() string {
	if u.Username != "" && u.Username != u.id {
		return fmt.Sprintf("%s: %s", u.id, u.Username)
	}
	return u.id
}
func (u *User) SetID(id string)  { u.id = id }
func (u *User) SetSkip()         { u.Skip = true }
func (u *User) TypeKey() string  { return UserResourceKey }
func (u *User) TypeName() string { return UserResourceName }
func (u *User) GetTitle() string { return u.Title }
func (u *User) GetMeta() meta    { return u.Meta }
func (u *User) GetUsername() string {
	if u.Username != "" {
		return u.Username
	}
	return u.id
}

func (u *User) Validate(sys *system.System) []TestResult {
	ctx := context.Background()
	skip := u.Skip
	sysuser := sys.NewUser(u.GetUsername(), sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(ctx, u, "exists", u.Exists, sysuser.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if u.UID != nil {
		uUID := deprecateAtoI(u.UID, fmt.Sprintf("%s: user.uid", u.Username))
		results = append(results, ValidateValue(ctx, u, "uid", uUID, sysuser.UID, skip))
	}
	if u.GID != nil {
		uGID := deprecateAtoI(u.GID, fmt.Sprintf("%s: user.gid", u.Username))
		results = append(results, ValidateValue(ctx, u, "gid", uGID, sysuser.GID, skip))
	}
	if u.Home != nil {
		results = append(results, ValidateValue(ctx, u, "home", u.Home, sysuser.Home, skip))
	}
	if u.Groups != nil {
		results = append(results, ValidateValue(ctx, u, "groups", u.Groups, sysuser.Groups, skip))
	}
	if u.Shell != nil {
		results = append(results, ValidateValue(ctx, u, "shell", u.Shell, sysuser.Shell, skip))
	}
	return results
}

func NewUser(sysUser system.User, config util.Config) (*User, error) {
	ctx := context.Background()
	username := sysUser.Username()
	exists, _ := sysUser.Exists(ctx)
	u := &User{
		id:     username,
		Exists: exists,
	}
	if !contains(config.IgnoreList, "uid") {
		if uid, err := sysUser.UID(ctx); err == nil {
			u.UID = uid
		}
	}
	if !contains(config.IgnoreList, "gid") {
		if gid, err := sysUser.GID(ctx); err == nil {
			u.GID = gid
		}
	}
	if !contains(config.IgnoreList, "groups") {
		if groups, err := sysUser.Groups(ctx); err == nil {
			u.Groups = groups
		}
	}
	if !contains(config.IgnoreList, "home") {
		if home, err := sysUser.Home(ctx); err == nil {
			u.Home = home
		}
	}
	if !contains(config.IgnoreList, "shell") {
		if shell, err := sysUser.Shell(ctx); err == nil {
			u.Shell = shell
		}
	}
	return u, nil
}
