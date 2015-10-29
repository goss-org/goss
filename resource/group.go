package resource

import "github.com/aelsabbahy/goss/system"

type Group struct {
	Groupname string `json:"-"`
	Exists    bool   `json:"exists"`
	Gid       string `json:"gid,omitempty"`
}

func (g *Group) ID() string      { return g.Groupname }
func (g *Group) SetID(id string) { g.Groupname = id }

func (g *Group) Validate(sys *system.System) []TestResult {
	sysgroup := sys.NewGroup(g.Groupname, sys)

	var results []TestResult

	results = append(results, ValidateValue(g, "exists", g.Exists, sysgroup.Exists))

	if g.Gid != "" {
		results = append(results, ValidateValue(g, "gid", g.Gid, sysgroup.Gid))
	}

	return results
}

func NewGroup(sysGroup system.Group, ignoreList []string) *Group {
	groupname := sysGroup.Groupname()
	exists, _ := sysGroup.Exists()
	g := &Group{
		Groupname: groupname,
		Exists:    exists.(bool),
	}
	if !contains(ignoreList, "stderr") {
		gid, _ := sysGroup.Gid()
		g.Gid = gid.(string)
	}
	return g
}
