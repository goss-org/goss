package system

import (
	"context"
	"errors"
	"strings"

	"github.com/goss-org/goss/util"
)

type DebPackage struct {
	name      string
	versions  []string
	loaded    bool
	installed bool
}

func NewDebPackage(_ context.Context, name string, system *System, config util.Config) Package {
	return &DebPackage{name: name}
}

func (p *DebPackage) setup() {
	if p.loaded {
		return
	}
	p.loaded = true
	cmd := util.NewCommand("dpkg-query", "-f", "${Status} ${Version}\n", "-W", p.name)
	if err := cmd.Run(); err != nil {
		return
	}
	for _, l := range strings.Split(strings.TrimSpace(cmd.Stdout.String()), "\n") {
		if !(strings.HasPrefix(l, "install ok installed") || strings.HasPrefix(l, "hold ok installed")) {
			continue
		}
		ver := strings.Fields(l)[3]
		p.versions = append(p.versions, ver)
	}

	if len(p.versions) > 0 {
		p.installed = true
	}
}

func (p *DebPackage) Name() string {
	return p.name
}

func (p *DebPackage) Exists() (bool, error) { return p.Installed() }

func (p *DebPackage) Installed() (bool, error) {
	p.setup()

	return p.installed, nil
}

func (p *DebPackage) Versions() ([]string, error) {
	p.setup()
	if len(p.versions) == 0 {
		return p.versions, errors.New("Package version not found")
	}
	return p.versions, nil
}
