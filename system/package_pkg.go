package system

import (
	"errors"
	"strings"

	"github.com/aelsabbahy/goss/util"
)

type PkgPackage struct {
	name      string
	versions  []string
	loaded    bool
	installed bool
}

func NewPkgPackage(name string, system *System, config util.Config) Package {
	return &PkgPackage{name: name}
}

func (p *PkgPackage) setup() {
	if p.loaded {
		return
	}
	p.loaded = true
	cmd := util.NewCommand("pkg", "query",
		"%v", p.name)
	if err := cmd.Run(); err != nil {
		return
	}
	p.installed = true
	p.versions = strings.Split(strings.TrimSpace(cmd.Stdout.String()), "\n")
}

func (p *PkgPackage) Name() string {
	return p.name
}

func (p *PkgPackage) Exists() (bool, error) { return p.Installed() }

func (p *PkgPackage) Installed() (bool, error) {
	p.setup()

	return p.installed, nil
}

func (p *PkgPackage) Versions() ([]string, error) {
	p.setup()
	if len(p.versions) == 0 {
		return p.versions, errors.New("Package version not found")
	}
	return p.versions, nil
}
