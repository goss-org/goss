package system

import "errors"

type Package interface {
	Name() string
	Exists() (interface{}, error)
	Installed() (interface{}, error)
	Versions() ([]string, error)
}

var ErrNullPackage = errors.New("Could not detect Package type on this system, please use --package flag to explicity set it")

type PackageNull struct {
	name string
}

func NewPackageNull(name string, system *System) Package {
	return &PackageNull{name: name}
}

func (p *PackageNull) Name() string { return p.name }

func (p *PackageNull) Exists() (interface{}, error) { return p.Installed() }

func (p *PackageNull) Installed() (interface{}, error) {
	return false, ErrNullPackage
}

func (p *PackageNull) Versions() ([]string, error) {
	return nil, ErrNullPackage
}
