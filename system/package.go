package system

import (
	"context"
	"errors"
	"fmt"

	"github.com/goss-org/goss/util"
)

type Package interface {
	Name() string
	Exists() (bool, error)
	Installed() (bool, error)
	Versions() ([]string, error)
}

var ErrNullPackage = errors.New("Could not detect Package type on this system, please use --package flag to explicity set it")

type NullPackage struct {
	name string
}

func NewNullPackage(_ context.Context, name interface{}, system *System, config util.Config) (Package, error) {
	strName, ok := name.(string)
	if !ok {
		return nil, fmt.Errorf("name must be of type string")
	}
	return newNullPackage(nil, strName, system, config), nil
}

func newNullPackage(_ context.Context, name string, system *System, config util.Config) Package {
	return &NullPackage{name: name}
}

func (p *NullPackage) Name() string { return p.name }

func (p *NullPackage) Exists() (bool, error) { return p.Installed() }

func (p *NullPackage) Installed() (bool, error) {
	return false, ErrNullPackage
}

func (p *NullPackage) Versions() ([]string, error) {
	return nil, ErrNullPackage
}
