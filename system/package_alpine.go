package system

import (
	"context"
	"errors"
	"strings"

	"github.com/goss-org/goss/util"
)

type AlpinePackage struct {
	name      string
	versions  []string
	loaded    bool
	installed bool
}

func NewAlpinePackage(name string, system *System, config util.Config) Package {
	return &AlpinePackage{name: name}
}

func (p *AlpinePackage) setup() {
	if p.loaded {
		return
	}
	p.loaded = true
	cmd := util.NewCommand("apk", "version", p.name)
	if err := cmd.Run(); err != nil {
		return
	}
	for _, l := range strings.Split(strings.TrimSpace(cmd.Stdout.String()), "\n") {
		if strings.HasPrefix(l, "Installed:") || strings.HasPrefix(l, "WARNING") {
			continue
		}
		ver := strings.TrimPrefix(strings.Fields(l)[0], p.name+"-")
		p.versions = append(p.versions, ver)
	}

	if len(p.versions) > 0 {
		p.installed = true
	}
}

func (p *AlpinePackage) Name() string {
	return p.name
}

func (p *AlpinePackage) Exists(ctx context.Context) (bool, error) { return p.Installed(ctx) }

func (p *AlpinePackage) Installed(ctx context.Context) (bool, error) {
	p.setup()

	return p.installed, nil
}

func (p *AlpinePackage) Versions(ctx context.Context) ([]string, error) {
	p.setup()
	if len(p.versions) == 0 {
		return p.versions, errors.New("Package version not found")
	}
	return p.versions, nil
}
