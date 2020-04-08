package system

import (
	"errors"

	"github.com/aelsabbahy/goss/util"
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

func NewNullPackage(name string, system *System, config util.Config) (Package, error) {
	return &NullPackage{name: name}, nil
}

func (p *NullPackage) Name() string { return p.name }

func (p *NullPackage) Exists() (bool, error) { return p.Installed() }

func (p *NullPackage) Installed() (bool, error) {
	return false, ErrNullPackage
}

func (p *NullPackage) Versions() ([]string, error) {
	return nil, ErrNullPackage
}
