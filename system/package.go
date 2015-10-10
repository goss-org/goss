package system

import "errors"

type Package interface {
	Name() string
	Exists() (interface{}, error)
	Installed() (interface{}, error)
	Versions() ([]string, error)
}

var ErrNullPackage = errors.New("Could not detect Package type on this system, please use --package flag to explicity set it")

type NullPackage struct {
	name string
}

func NewNullPackage(name string, system *System) Package {
	return &NullPackage{name: name}
}

func (p *NullPackage) Name() string { return p.name }

func (p *NullPackage) Exists() (interface{}, error) { return p.Installed() }

func (p *NullPackage) Installed() (interface{}, error) {
	return false, ErrNullPackage
}

func (p *NullPackage) Versions() ([]string, error) {
	return nil, ErrNullPackage
}
