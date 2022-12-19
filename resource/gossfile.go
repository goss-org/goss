package resource

import (
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Gossfile struct {
	Title string `json:"title,omitempty" yaml:"title,omitempty"`
	Meta  meta   `json:"meta,omitempty" yaml:"meta,omitempty"`
	Path  string `json:"-" yaml:"-"`
	Skip  bool   `json:"skip,omitempty" yaml:"skip,omitempty"`
	File  string `json:"file,omitempty" yaml:"file,omitempty"`
}

const (
	GossFileResourceKey  = "gossfile"
	GossFileResourceName = "Gossfile"
)

func init() {
	registerResource(GossFileResourceKey, &Gossfile{})
}

func (g *Gossfile) ID() string       { return g.Path }
func (g *Gossfile) SetID(id string)  { g.Path = id }
func (g *Gossfile) SetSkip()         {}
func (g *Gossfile) TypeKey() string  { return GossFileResourceKey }
func (g *Gossfile) TypeName() string { return GossFileResourceName }

func (g *Gossfile) GetTitle() string { return g.Title }
func (g *Gossfile) GetMeta() meta    { return g.Meta }

func (g *Gossfile) GetSkip() bool { return g.Skip }

func (g *Gossfile) GetGossfile() string {
	if g.File != "" {
		return g.File
	}
	return g.Path
}

func (g *Gossfile) Validate(sys *system.System) []TestResult {
	return []TestResult{}
}

func NewGossfile(sysGossfile system.Gossfile, config util.Config) (*Gossfile, error) {
	path := sysGossfile.Path()
	return &Gossfile{
		Path: path,
	}, nil
}
