package resource

import (
	"fmt"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Group struct {
	Title     string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta      meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Groupname string  `json:"-" yaml:"-"`
	Exists    matcher `json:"exists" yaml:"exists"`
	GID       matcher `json:"gid,omitempty" yaml:"gid,omitempty"`
	Skip      bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	GroupResourceKey  = "group"
	GroupResourceName = "Group"
)

func init() {
	registerResource(GroupResourceKey, &Group{})
}

func (g *Group) ID() string       { return g.Groupname }
func (g *Group) SetID(id string)  { g.Groupname = id }
func (g *Group) SetSkip()         { g.Skip = true }
func (g *Group) TypeKey() string  { return GroupResourceKey }
func (g *Group) TypeName() string { return GroupResourceName }
func (g *Group) GetTitle() string { return g.Title }
func (g *Group) GetMeta() meta    { return g.Meta }

func (g *Group) Validate(sys *system.System) []TestResult {
	skip := g.Skip
	sysgroup := sys.NewGroup(g.Groupname, sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(g, "exists", g.Exists, sysgroup.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if g.GID != nil {
		gGID := deprecateAtoI(g.GID, fmt.Sprintf("%s: group.gid", g.Groupname))
		results = append(results, ValidateValue(g, "gid", gGID, sysgroup.GID, skip))
	}
	return results
}

func NewGroup(sysGroup system.Group, config util.Config) (*Group, error) {
	groupname := sysGroup.Groupname()
	exists, _ := sysGroup.Exists()
	g := &Group{
		Groupname: groupname,
		Exists:    exists,
	}
	if !contains(config.IgnoreList, "stderr") {
		if gid, err := sysGroup.GID(); err == nil {
			g.GID = gid
		}
	}
	return g, nil
}
