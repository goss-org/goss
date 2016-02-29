package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Package struct {
	Title     string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta      meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Name      string  `json:"-" yaml:"-"`
	Installed bool    `json:"installed" yaml:"installed"`
	Versions  matcher `json:"versions,omitempty" yaml:"versions,omitempty"`
}

func (p *Package) ID() string      { return p.Name }
func (p *Package) SetID(id string) { p.Name = id }

func (p *Package) GetTitle() string { return p.Title }
func (p *Package) GetMeta() meta    { return p.Meta }

func (p *Package) Validate(sys *system.System) []TestResult {
	sysPkg := sys.NewPackage(p.Name, sys, util.Config{})

	var results []TestResult

	results = append(results, ValidateValue(p, "installed", p.Installed, sysPkg.Installed))

	if p.Versions != nil {
		results = append(results, ValidateValue(p, "version", p.Versions, sysPkg.Versions))
	}

	return results
}

func NewPackage(sysPackage system.Package, config util.Config) (*Package, error) {
	name := sysPackage.Name()
	installed, _ := sysPackage.Installed()
	p := &Package{
		Name:      name,
		Installed: installed,
	}
	if !contains(config.IgnoreList, "versions") {
		if versions, err := sysPackage.Versions(); err == nil {
			p.Versions = versions
		}
	}
	return p, nil
}
