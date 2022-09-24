package resource

import (
	"fmt"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Process struct {
	Title   string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta    meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id      string  `json:"-" yaml:"-"`
	Comm    string  `json:"comm,omitempty" yaml:"comm,omitempty"`
	Running matcher `json:"running" yaml:"running"`
	Skip    bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

func (p *Process) ID() string {
	if p.Comm != "" && p.Comm != p.id {
		return fmt.Sprintf("%s: %s", p.id, p.Comm)
	}
	return p.id
}
func (p *Process) SetID(id string) { p.id = id }

func (p *Process) GetTitle() string { return p.Title }
func (p *Process) GetMeta() meta    { return p.Meta }
func (p *Process) GetComm() string {
	if p.Comm != "" {
		return p.Comm
	}
	return p.id
}

func (p *Process) Validate(sys *system.System) []TestResult {
	skip := false
	sysProcess := sys.NewProcess(p.GetComm(), sys, util.Config{})

	if p.Skip {
		skip = true
	}

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
