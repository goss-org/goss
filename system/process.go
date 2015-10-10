package system

import (
	"fmt"
	"os"
	"sync"

	"github.com/mitchellh/go-ps"
)

type Process interface {
	Executable() string
	Exists() (interface{}, error)
	Running() (interface{}, error)
	Pids() ([]int, error)
}

type DefProcess struct {
	executable string
}

// FIXME: eww
var processOnce sync.Once
var pmap map[string][]ps.Process

func initProcesses() {
	pmap = make(map[string][]ps.Process)
	processes, err := ps.Processes()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, p := range processes {
		pmap[p.Executable()] = append(pmap[p.Executable()], p)
	}
}

func NewDefProcess(executable string, system *System) Process {
	processOnce.Do(initProcesses)
	return &DefProcess{executable: executable}
}

func (p *DefProcess) Executable() string {
	return p.executable
}

func (p *DefProcess) Exists() (interface{}, error) { return p.Running() }

func (p *DefProcess) Pids() ([]int, error) {
	var pids []int
	for _, proc := range pmap[p.executable] {
		pids = append(pids, proc.Pid())
	}
	//if proc, ok := pmap[p.executable]; ok {

	//	return proc.Pid(), nil
	//}
	return pids, nil
}

func (p *DefProcess) Running() (interface{}, error) {
	if _, ok := pmap[p.executable]; ok {
		return true, nil
	}
	return false, nil
}
