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

func NewPackage(sysPackage system.Package, ignoreList []string) *Package {
	name := sysPackage.Name()
	installed, _ := sysPackage.Installed()
	p := &Package{
		Name:      name,
		Installed: installed.(bool),
	}
	if !contains(ignoreList, "versions") {
		versions, _ := sysPackage.Versions()
		p.Versions = versions
	}
	return p
}
