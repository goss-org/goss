package system

import (
	"errors"
	"strings"

	"github.com/aelsabbahy/goss/util"
)

type PackageRpm struct {
	name      string
	versions  []string
	loaded    bool
	installed bool
}

func NewPackageRpm(name string, system *System) Package {
	return &PackageRpm{name: name}
}

func (p *PackageRpm) setup() {
	if p.loaded {
		return
	}
	p.loaded = true
	cmd := util.NewCommand("rpm", "-q", "--nosignature", "--nohdrchk", "--nodigest", "--qf", "%{VERSION}\n", p.name)
	if err := cmd.Run(); err != nil {
		return
	}
	p.installed = true
	p.versions = strings.Split(strings.TrimSpace(cmd.Stdout.String()), "\n")
}

func (p *PackageRpm) Name() string {
	return p.name
}

func (p *PackageRpm) Installed() (interface{}, error) {
	p.setup()

	return p.installed, nil
}

func (p *PackageRpm) Versions() ([]string, error) {
	p.setup()
	if len(p.versions) == 0 {
		return p.versions, errors.New("Package version not found")
	}
	return p.versions, nil
}
