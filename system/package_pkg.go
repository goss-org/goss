package system

import (
	"errors"
	"strings"

	"github.com/aelsabbahy/goss/util"
)

// PkgPackage is FreeBSD pkg(8) structure.
type PkgPackage struct {
	name      string
	versions  []string
	loaded    bool
	installed bool
}

// NewPkgPackage returns interface to test given package information.
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

// Name returns package name.
func (p *PkgPackage) Name() string {
	return p.name
}

// Exists returns boolean that given package is installed or not.
func (p *PkgPackage) Exists() (bool, error) { return p.Installed() }

// Installed is same as Exists.
func (p *PkgPackage) Installed() (bool, error) {
	p.setup()

	return p.installed, nil
}

// Versions returns given package versions.
func (p *PkgPackage) Versions() ([]string, error) {
	p.setup()
	if len(p.versions) == 0 {
		return p.versions, errors.New("Package version not found")
	}
	return p.versions, nil
}
