package system

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/util"
)

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

func NewDefGossfile(_ context.Context, path interface{}, system *System, config util.Config) (Gossfile, error) {
	strPath, ok := path.(string)
	if !ok {
		return nil, fmt.Errorf("path must be of type string")
	}
	return newDefGossfile(nil, strPath, system, config), nil
}

func newDefGossfile(_ context.Context, path string, system *System, config util.Config) Gossfile {
	return &DefGossfile{path: path}
}
