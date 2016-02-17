package system

import "github.com/aelsabbahy/goss/util"

type Gossfile interface {
	Path() string
	Exists() (bool, error)
}

type DefGossfile struct {
	path string
}

func (g *DefGossfile) Path() string {
	return g.path
}

// Stub out
func (g *DefGossfile) Exists() (bool, error) {
	return false, nil
}

func NewDefGossfile(path string, system *System, config util.Config) Gossfile {
	return &DefGossfile{path: path}
}
