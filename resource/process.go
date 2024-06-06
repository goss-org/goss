package resource

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Process struct {
	Title   string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta    meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id      string  `json:"-" yaml:"-"`
	Comm    string  `json:"comm,omitempty" yaml:"comm,omitempty"`
	Running matcher `json:"running" yaml:"running"`
	Skip    bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	ProcessResourceKey  = "process"
	ProcessResourceName = "Process"
)

func init() {
	registerResource(ProcessResourceKey, &Process{})
}

func (p *Process) ID() string {
	if p.Comm != "" && p.Comm != p.id {
		return fmt.Sprintf("%s: %s", p.id, p.Comm)
	}
	return p.id
}
func (p *Process) SetID(id string)  { p.id = id }
func (p *Process) SetSkip()         { p.Skip = true }
func (p *Process) TypeKey() string  { return ProcessResourceKey }
func (p *Process) TypeName() string { return ProcessResourceName }
func (p *Process) GetTitle() string { return p.Title }
func (p *Process) GetMeta() meta    { return p.Meta }
func (p *Process) GetComm() string {
	if p.Comm != "" {
		return p.Comm
	}
	return p.id
}

func (p *Process) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), idKey{}, p.ID())
	skip := p.Skip
	sysProcess := sys.NewProcess(ctx, p.GetComm(), sys, util.Config{})

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
		id:      executable,
		Running: running,
	}, nil
}
