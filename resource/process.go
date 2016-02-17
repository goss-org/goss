package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Process struct {
	Desc       string `json:"desc,omitempty" yaml:"desc,omitempty"`
	Executable string `json:"-" yaml:"-"`
	Running    bool   `json:"running" yaml:"running"`
}

func (p *Process) ID() string      { return p.Executable }
func (p *Process) SetID(id string) { p.Executable = id }

func (p *Process) Validate(sys *system.System) []TestResult {
	sysProcess := sys.NewProcess(p.Executable, sys, util.Config{})

	var results []TestResult

	results = append(results, ValidateValue(p, "running", p.Running, sysProcess.Running))

	return results
}

func NewProcess(sysProcess system.Process, config util.Config) (*Process, error) {
	executable := sysProcess.Executable()
	running, _ := sysProcess.Running()
	return &Process{
		Executable: executable,
		Running:    running,
	}, nil
}
