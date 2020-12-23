package system

import (
	"github.com/aelsabbahy/goss/util"
	"github.com/shirou/gopsutil/process"
)

type Process interface {
	Executable() string
	Exists() (bool, error)
	Running() (bool, error)
	Pids() ([]int, error)
}

type DefProcess struct {
	executable string
	procMap    map[string][]*process.Process
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
		pids = append(pids, int(proc.Pid))
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

func GetProcs() (map[string][]*process.Process, error) {
	pmap := make(map[string][]*process.Process)
	processes, err := process.Processes()
	if err != nil {
		return pmap, err
	}
	for _, p := range processes {
		if pExe, err := p.Name(); err == nil {
			pmap[pExe] = append(pmap[pExe], p)
		}
	}

	return pmap, nil
}
