package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Gossfile struct {
	Title string `json:"title,omitempty" yaml:"title,omitempty"`
	Meta  meta   `json:"meta,omitempty" yaml:"meta,omitempty"`
	Path  string `json:"-" yaml:"-"`
}

const (
	GossFileResourceKey  = "gossfile"
	GossFileResourceName = "Gossfile"
)

func init() {
	registerResource(GossFileResourceKey, &Gossfile{})
}

func (g *Gossfile) ID() string      { return g.Path }
func (g *Gossfile) SetID(id string) { g.Path = id }

func (g *Gossfile) GetTitle() string { return g.Title }
func (g *Gossfile) GetMeta() meta    { return g.Meta }

func (g *Gossfile) Validate(sys *system.System, skipTypes []string) []TestResult {
	return []TestResult{}
}

func NewGossfile(sysGossfile system.Gossfile, config util.Config) (*Gossfile, error) {
	path := sysGossfile.Path()
	return &Gossfile{
		Path: path,
	}, nil
}
