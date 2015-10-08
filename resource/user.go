package resource

import "github.com/aelsabbahy/goss/system"

type User struct {
	Username string   `json:"username"`
	Exists   bool     `json:"exists"`
	UID      string   `json:"uid,omitempty"`
	GID      string   `json:"gid,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Home     string   `json:"home,omitempty"`
}

func (s *User) Validate(sys *system.System) []TestResult {
	sysuser := sys.NewUser(s.Username, sys)

	var results []TestResult

	results = append(results, ValidateValue(s.Username, "exists", s.Exists, sysuser.Exists))
	if !s.Exists {
		return results
	}
	results = append(results, ValidateValue(s.Username, "uid", s.UID, sysuser.UID))
	results = append(results, ValidateValue(s.Username, "gid", s.GID, sysuser.GID))
	results = append(results, ValidateValue(s.Username, "home", s.Home, sysuser.Home))
	results = append(results, ValidateValues(s.Username, "groups", s.Groups, sysuser.Groups))

	return results
}

func NewUser(sysUser system.User) *User {
	username := sysUser.Username()
	exists, _ := sysUser.Exists()
	uid, _ := sysUser.UID()
	gid, _ := sysUser.GID()
	groups, _ := sysUser.Groups()
	home, _ := sysUser.Home()
	return &User{
		Username: username,
		Exists:   exists.(bool),
		UID:      uid.(string),
		GID:      gid.(string),
		Groups:   groups,
		Home:     home.(string),
	}
}
