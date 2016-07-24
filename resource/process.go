package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Process struct {
	Title      string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Executable string  `json:"-" yaml:"-"`
	Running    matcher `json:"running" yaml:"running"`
}

func (p *Process) ID() string      { return p.Executable }
func (p *Process) SetID(id string) { p.Executable = id }

func (p *Process) GetTitle() string { return p.Title }
func (p *Process) GetMeta() meta    { return p.Meta }

func (p *Process) Validate(sys *system.System) []TestResult {
	skip := false
	sysProcess := sys.NewProcess(p.Executable, sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(p, "running", p.Running, sysProcess.Running, skip))
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
