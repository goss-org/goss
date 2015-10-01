package resource

import "github.com/aelsabbahy/goss/system"

type Group struct {
	Groupname string `json:"groupname"`
	Exists    bool   `json:"exists"`
	Gid       string `json:"gid,omitempty"`
}

func (s *Group) Validate(sys *system.System) []TestResult {
	sysgroup := sys.NewGroup(s.Groupname, sys)

	var results []TestResult

	results = append(results, ValidateValue(s.Groupname, "exists", s.Exists, sysgroup.Exists))
	if !s.Exists {
		return results
	}
	results = append(results, ValidateValue(s.Gid, "gid", s.Gid, sysgroup.Gid))

	return results
}

func NewGroup(sysGroup system.Group) *Group {
	groupname := sysGroup.Groupname()
	exists, _ := sysGroup.Exists()
	gid, _ := sysGroup.Gid()
	return &Group{
		Groupname: groupname,
		Exists:    exists.(bool),
		Gid:       gid.(string),
	}
}
