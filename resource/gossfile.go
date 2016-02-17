package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Gossfile struct {
	Desc string `json:"desc,omitempty" yaml:"desc,omitempty"`
	Path string `json:"-" yaml:"-"`
}

func (g *Gossfile) ID() string      { return g.Path }
func (g *Gossfile) SetID(id string) { g.Path = id }

func NewGossfile(sysGossfile system.Gossfile, config util.Config) (*Gossfile, error) {
	path := sysGossfile.Path()
	return &Gossfile{
		Path: path,
	}, nil
}
