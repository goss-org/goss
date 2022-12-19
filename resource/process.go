package resource

import (
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Process struct {
	Title      string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Executable string  `json:"-" yaml:"-"`
	Running    matcher `json:"running" yaml:"running"`
	Skip       bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	ProcessResourceKey  = "process"
	ProcessResourceName = "Process"
)

func init() {
	registerResource(ProcessResourceKey, &Process{})
}

func (p *Process) ID() string       { return p.Executable }
func (p *Process) SetID(id string)  { p.Executable = id }
func (p *Process) SetSkip()         { p.Skip = true }
func (p *Process) TypeKey() string  { return ProcessResourceKey }
func (p *Process) TypeName() string { return ProcessResourceName }
func (p *Process) GetTitle() string { return p.Title }
func (p *Process) GetMeta() meta    { return p.Meta }

func (p *Process) Validate(sys *system.System) []TestResult {
	skip := p.Skip
	sysProcess := sys.NewProcess(p.Executable, sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(p, "running", p.Running, sysProcess.Running, skip))
	return results
}

func NewProcess(sysProcess system.Process, config util.Config) (*Process, error) {
	executable := sysProcess.Executable()
	running, err := sysProcess.Running()
	if err != nil {
		return nil, err
	}
	return &Process{
		Executable: executable,
		Running:    running,
	}, nil
}
