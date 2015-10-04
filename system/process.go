package system

import (
	"fmt"
	"os"
	"sync"

	"github.com/mitchellh/go-ps"
)

type Process struct {
	executable string
	running    bool
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

func NewProcess(executable string, system *System) *Process {
	processOnce.Do(initProcesses)
	return &Process{executable: executable}
}

func (p *Process) Executable() string {
	return p.executable
}

func (p *Process) Exists() (interface{}, error) { return p.Running() }

func (p *Process) Pids() ([]int, error) {
	var pids []int
	for _, proc := range pmap[p.executable] {
		pids = append(pids, proc.Pid())
	}
	//if proc, ok := pmap[p.executable]; ok {

	//	return proc.Pid(), nil
	//}
	return pids, nil
}

func (p *Process) Running() (interface{}, error) {
	if _, ok := pmap[p.executable]; ok {
		return true, nil
	}
	return false, nil
}
