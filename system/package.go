package system

import (
	"context"
	"errors"

	"github.com/goss-org/goss/util"
)

type Package interface {
	Name() string
	Exists(context.Context) (bool, error)
	Installed(context.Context) (bool, error)
	Versions(context.Context) ([]string, error)
}

var ErrNullPackage = errors.New("Could not detect Package type on this system, please use --package flag to explicity set it")

type NullPackage struct {
	name string
}

func NewNullPackage(name string, system *System, config util.Config) Package {
	return &NullPackage{name: name}
}

func (p *NullPackage) Name() string { return p.name }

func (p *NullPackage) Exists(ctx context.Context) (bool, error) { return p.Installed(ctx) }

func (p *NullPackage) Installed(ctx context.Context) (bool, error) {
	return false, ErrNullPackage
}

func (p *NullPackage) Versions(ctx context.Context) ([]string, error) {
	return nil, ErrNullPackage
}
