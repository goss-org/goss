package system

import (
	"fmt"
	"os"

	"github.com/aelsabbahy/goss/util"
	"github.com/mitchellh/go-ps"
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
}

func NewDefProcess(executable string, system *System, config util.Config) Process {
	return &DefProcess{
		executable: executable,
		procMap:    system.ProcMap(),
	}
}

func (p *DefProcess) Executable() string {
	return p.executable
}

func (p *DefProcess) Exists() (bool, error) { return p.Running() }

func (p *DefProcess) Pids() ([]int, error) {
	var pids []int
	for _, proc := range p.procMap[p.executable] {
		pids = append(pids, proc.Pid())
	}
	return pids, nil
}

func (p *DefProcess) Running() (bool, error) {
	if _, ok := p.procMap[p.executable]; ok {
		return true, nil
	}
	return false, nil
}

func GetProcs() map[string][]ps.Process {
	pmap := make(map[string][]ps.Process)
	processes, err := ps.Processes()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, p := range processes {
		pmap[p.Executable()] = append(pmap[p.Executable()], p)
	}

	return pmap
}
