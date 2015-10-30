package resource

import "github.com/aelsabbahy/goss/system"

type Gossfile struct {
	Path string `json:"-"`
}

func (g *Gossfile) ID() string      { return g.Path }
func (g *Gossfile) SetID(id string) { g.Path = id }

func NewGossfile(sysGossfile system.Gossfile, ignoreList []string) *Gossfile {
	path := sysGossfile.Path()
	return &Gossfile{
		Path: path,
	}
}
