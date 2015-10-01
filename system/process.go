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
var pmap map[string]bool

func initProcesses() {
	pmap = make(map[string]bool)
	processes, err := ps.Processes()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, p := range processes {
		pmap[p.Executable()] = true
	}
}

func NewProcess(executable string, system *System) *Process {
	processOnce.Do(initProcesses)
	return &Process{executable: executable}
}

func (p *Process) Executable() string {
	return p.executable
}

func (p *Process) Running() (interface{}, error) {
	if pmap[p.executable] {
		return true, nil
	}
	return false, nil
}
