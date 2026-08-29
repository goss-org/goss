package util

import (
	"bytes"
	"encoding/json"

	//"fmt"
	"errors"
	"os/exec"
	"syscall"
)

// ExecCommand represents a command that can be specified either shell style,
// as a single string that is run through the shell (`sh -c "<string>"`), or
// exec style, as an explicit list of arguments where the first element is the
// program and the rest are its arguments (no shell involved). This mirrors the
// shell/exec forms of Dockerfile's RUN/ENTRYPOINT/CMD instructions.
//
// Exactly one of CmdStr or CmdSlice is set. It is useful for environments
// without a shell, such as distroless/scratch containers, and for passing
// arguments containing spaces or special characters verbatim.
type ExecCommand struct {
	CmdStr   string
	CmdSlice []string
}

// errExecCommandType is returned when a command value is neither a string
// (shell style) nor a list of strings (exec style).
var errExecCommandType = errors.New("command must be a string or a list of strings")

// UnmarshalJSON accepts either a JSON string (shell style) or a JSON array of
// strings (exec style). Anything else is an error.
func (e *ExecCommand) UnmarshalJSON(data []byte) error {
	// Try shell style first.
	if err := json.Unmarshal(data, &e.CmdStr); err == nil {
		return nil
	}
	// Fall back to exec style.
	e.CmdStr = ""
	if err := json.Unmarshal(data, &e.CmdSlice); err != nil {
		return errExecCommandType
	}
	return nil
}

// UnmarshalYAML accepts either a YAML scalar (shell style) or a YAML sequence
// of strings (exec style). Anything else is an error.
func (e *ExecCommand) UnmarshalYAML(unmarshal func(any) error) error {
	// Try shell style first. A bool/int scalar decodes into the string, which
	// preserves the long-standing behavior of `exec: true` meaning `exec: "true"`.
	if err := unmarshal(&e.CmdStr); err == nil {
		return nil
	}
	// Fall back to exec style.
	e.CmdStr = ""
	if err := unmarshal(&e.CmdSlice); err != nil {
		return errExecCommandType
	}
	return nil
}

// MarshalJSON emits the command as a JSON string (shell style) or array of
// strings (exec style), matching what UnmarshalJSON accepts.
func (e ExecCommand) MarshalJSON() ([]byte, error) {
	if e.CmdStr != "" {
		return json.Marshal(e.CmdStr)
	}
	return json.Marshal(e.CmdSlice)
}

// MarshalYAML emits the command as a YAML scalar (shell style) or sequence of
// strings (exec style), matching what UnmarshalYAML accepts.
func (e ExecCommand) MarshalYAML() (any, error) {
	if e.CmdStr != "" {
		return e.CmdStr, nil
	}
	return e.CmdSlice, nil
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
