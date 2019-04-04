package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Package struct {
	Title     string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta      meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Name      string  `json:"-" yaml:"-"`
	Installed matcher `json:"installed" yaml:"installed"`
	Versions  matcher `json:"versions,omitempty" yaml:"versions,omitempty"`
	Skip      bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

func (p *Package) ID() string      { return p.Name }
func (p *Package) SetID(id string) { p.Name = id }

func (p *Package) GetTitle() string { return p.Title }
func (p *Package) GetMeta() meta    { return p.Meta }

func (p *Package) Validate(sys *system.System) []TestResult {
	skip := false
	sysPkg := sys.NewPackage(p.Name, sys, util.Config{})

	if p.Skip { skip = true }

	var results []TestResult
	results = append(results, ValidateValue(p, "installed", p.Installed, sysPkg.Installed, skip))
	if shouldSkip(results) {
		skip = true
	}
	if p.Versions != nil {
		results = append(results, ValidateValue(p, "version", p.Versions, sysPkg.Versions, skip))
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
