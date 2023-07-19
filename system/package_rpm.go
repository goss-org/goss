package system

import (
	"errors"
	"strings"

	"github.com/goss-org/goss/util"
)

type RpmPackage struct {
	name      string
	versions  []string
	loaded    bool
	installed bool
}

func NewRpmPackage(name string, system *System, config util.Config) Package {
	return &RpmPackage{name: name}
}

func (p *RpmPackage) setup() {
	if p.loaded {
		return
	}
	p.loaded = true
	cmd := util.NewCommand("rpm", "-q", "--nosignature", "--nohdrchk", "--nodigest", "--qf", "%|EPOCH?{%{EPOCH}:}:{}|%{VERSION}-%{RELEASE}\n", p.name)
	if err := cmd.Run(); err != nil {
		return
	}
	p.installed = true
	p.versions = strings.Split(strings.TrimSpace(cmd.Stdout.String()), "\n")
}

func (p *RpmPackage) Name() string {
	return p.name
}

func (p *RpmPackage) Exists() (bool, error) { return p.Installed() }

func (p *RpmPackage) Installed() (bool, error) {
	p.setup()

	return p.installed, nil
}

func (p *RpmPackage) Versions() ([]string, error) {
	p.setup()
	if len(p.versions) == 0 {
		return p.versions, errors.New("Package version not found")
	}
	return p.versions, nil
}
