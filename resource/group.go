package resource

import (
	"fmt"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Group struct {
	Title     string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta      meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id        string  `json:"-" yaml:"-"`
	Groupname string  `json:"groupname,omitempty" yaml:"groupname,omitempty"`
	Exists    matcher `json:"exists" yaml:"exists"`
	GID       matcher `json:"gid,omitempty" yaml:"gid,omitempty"`
	Skip      bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

func (g *Group) ID() string {
	if g.Groupname != "" && g.Groupname != g.id {
		return fmt.Sprintf("%s: %s", g.id, g.Groupname)
	}
	return g.id
}
func (g *Group) SetID(id string) { g.id = id }

func (g *Group) GetTitle() string { return g.Title }
func (g *Group) GetMeta() meta    { return g.Meta }
func (g *Group) GetGroupname() string {
	if g.Groupname != "" {
		return g.Groupname
	}
	return g.id
}

func (g *Group) Validate(sys *system.System) []TestResult {
	skip := false
	sysgroup := sys.NewGroup(g.GetGroupname(), sys, util.Config{})

	if g.Skip {
		skip = true
	}

	var results []TestResult
	results = append(results, ValidateValue(g, "exists", g.Exists, sysgroup.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if g.GID != nil {
		gGID := deprecateAtoI(g.GID, fmt.Sprintf("%s: group.gid", g.ID()))
		results = append(results, ValidateValue(g, "gid", gGID, sysgroup.GID, skip))
	}
	return results
}

func NewGroup(sysGroup system.Group, config util.Config) (*Group, error) {
	groupname := sysGroup.Groupname()
	exists, _ := sysGroup.Exists()
	g := &Group{
		id:     groupname,
		Exists: exists,
	}
	if !contains(config.IgnoreList, "stderr") {
		if gid, err := sysGroup.GID(); err == nil {
			g.GID = gid
		}
	}
	return g, nil
}
