package system

import (
	"github.com/goss-org/go-ps"
	"github.com/goss-org/goss/util"
)

type Process interface {
	Executable() string
	Exists() (bool, error)
	Running() (bool, error)
	Pids() ([]int, error)
}

type DefProcess struct {
	executable string
	procMap    map[string][]ps.Process
	err        error
}

func NewDefProcess(executable string, system *System, config util.Config) Process {
	pmap, err := system.ProcMap()
	return &DefProcess{
		executable: executable,
		procMap:    pmap,
		err:        err,
	}
}

func (p *DefProcess) Executable() string {
	return p.executable
}

func (p *DefProcess) Exists() (bool, error) { return p.Running() }

func (p *DefProcess) Pids() ([]int, error) {
	var pids []int
	if p.err != nil {
		return pids, p.err
	}
	for _, proc := range p.procMap[p.executable] {
		pids = append(pids, proc.Pid())
	}
	return pids, nil
}

func (p *DefProcess) Running() (bool, error) {
	if p.err != nil {
		return false, p.err
	}
	if _, ok := p.procMap[p.executable]; ok {
		return true, nil
	}
	return false, nil
}

func GetProcs() (map[string][]ps.Process, error) {
	pmap := make(map[string][]ps.Process)
	processes, err := ps.Processes()
	if err != nil {
		return pmap, err
	}
	for _, p := range processes {
		pmap[p.Executable()] = append(pmap[p.Executable()], p)
	}

	return pmap, nil
}
