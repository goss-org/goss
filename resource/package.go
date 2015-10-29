package resource

import "github.com/aelsabbahy/goss/system"

type Package struct {
	Name      string   `json:"-"`
	Installed bool     `json:"installed"`
	Versions  []string `json:"versions,omitempty"`
}

func (p *Package) ID() string      { return p.Name }
func (p *Package) SetID(id string) { p.Name = id }

func (p *Package) Validate(sys *system.System) []TestResult {
	sysPkg := sys.NewPackage(p.Name, sys)

	var results []TestResult

	results = append(results, ValidateValue(p, "installed", p.Installed, sysPkg.Installed))

	if len(p.Versions) > 0 {
		results = append(results, ValidateValues(p, "version", p.Versions, sysPkg.Versions))
	}

	return results
}

func NewPackage(sysPackage system.Package) *Package {
	name := sysPackage.Name()
	versions, _ := sysPackage.Versions()
	installed, _ := sysPackage.Installed()
	return &Package{
		Name:      name,
		Versions:  versions,
		Installed: installed.(bool),
	}
}
