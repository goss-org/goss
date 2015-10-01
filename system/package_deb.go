package system

import (
	"errors"
	"strings"

	"github.com/aelsabbahy/goss/util"
)

type PackageDeb struct {
	name      string
	versions  []string
	loaded    bool
	installed bool
}

func NewPackageDeb(name string, system *System) Package {
	return &PackageDeb{name: name}
}

func (p *PackageDeb) setup() {
	if p.loaded {
		return
	}
	p.loaded = true
	cmd := util.NewCommand("dpkg-query", "-f", "${Version}\n", "-W", p.name)
	if err := cmd.Run(); err != nil {
		return
	}
	p.installed = true
	p.versions = strings.Split(strings.TrimSpace(cmd.Stdout.String()), "\n")
}

func (p *PackageDeb) Name() string {
	return p.name
}

func (p *PackageDeb) Installed() (interface{}, error) {
	p.setup()

	return p.installed, nil
}

func (p *PackageDeb) Versions() ([]string, error) {
	p.setup()
	if len(p.versions) == 0 {
		return p.versions, errors.New("Package version not found")
	}
	return p.versions, nil
}
