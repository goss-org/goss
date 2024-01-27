package util

import (
	"bytes"
	"encoding/json"

	//"fmt"
	"os/exec"
	"syscall"
)

// Allows passing a shell style command string
// or an exec style slice of strings.
type ExecCommand struct {
	CmdStr   string
	CmdSlice []string
}

func (e *ExecCommand) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string
	if err := json.Unmarshal(data, &e.CmdStr); err != nil {
		// If string unmarshalling fails, try as a slice
		return json.Unmarshal(data, &e.CmdSlice)
	}
	return nil
}

func (e *ExecCommand) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as a string
	if err := unmarshal(&e.CmdStr); err != nil {
		// If string unmarshalling fails, try as a slice
		return unmarshal(&e.CmdSlice)
	}
	return nil
}

type Command struct {
	name           string
	Cmd            *exec.Cmd
	Stdout, Stderr bytes.Buffer
	Err            error
	Status         int
}

func NewCommand(name string, arg ...string) *Command {
	//fmt.Println(arg)
	command := new(Command)
	command.name = name
	command.Cmd = exec.Command(name, arg...)

	return command
}

func (c *Command) Run() error {
	c.Cmd.Stdout = &c.Stdout
	c.Cmd.Stderr = &c.Stderr

	if _, err := exec.LookPath(c.name); err != nil {
		c.Err = err
		return c.Err
	}

	if err := c.Cmd.Start(); err != nil {
		c.Err = err
		return c.Err
	}

	if err := c.Cmd.Wait(); err != nil {
		c.Err = err
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				c.Status = status.ExitStatus()
			}
		}
	} else {
		c.Status = 0
	}
	return c.Err
}
