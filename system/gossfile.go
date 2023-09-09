package system

import (
	"context"

	"github.com/goss-org/goss/util"
)

type Gossfile interface {
	Path() string
	Exists(context.Context) (bool, error)
}

type DefGossfile struct {
	path string
}

func (g *DefGossfile) Path() string {
	return g.path
}

// Stub out
func (g *DefGossfile) Exists(ctx context.Context) (bool, error) {
	return false, nil
}

func NewDefGossfile(path string, system *System, config util.Config) Gossfile {
	return &DefGossfile{path: path}
}
