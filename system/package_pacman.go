package system

import (
	"context"
	"errors"
	"strings"

	"github.com/goss-org/goss/util"
)

type PacmanPackage struct {
	name      string
	versions  []string
	loaded    bool
	installed bool
}

func NewPacmanPackage(name string, system *System, config util.Config) Package {
	return &PacmanPackage{name: name}
}

func (p *PacmanPackage) setup() {
	if p.loaded {
		return
	}
	p.loaded = true
	// TODO: extract versions
	cmd := util.NewCommand("pacman", "-Q", "--color", "never", "--noconfirm", p.name)
	if err := cmd.Run(); err != nil {
		return
	}
	p.installed = true
	// the output format is "pkgname version\n", so if we split the string on
	// whitespace, the version is the second item.
	p.versions = []string{strings.Fields(cmd.Stdout.String())[1]}
}

func (p *PacmanPackage) Name() string {
	return p.name
}

func (p *PacmanPackage) Exists(ctx context.Context) (bool, error) { return p.Installed(ctx) }

func (p *PacmanPackage) Installed(ctx context.Context) (bool, error) {
	p.setup()

	return p.installed, nil
}

func (p *PacmanPackage) Versions(ctx context.Context) ([]string, error) {
	p.setup()
	if len(p.versions) == 0 {
		return p.versions, errors.New("Package version not found")
	}
	return p.versions, nil
}
