package system

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/goss-org/goss/util"
)

type Command interface {
	Command() util.ExecCommand
	Exists() (bool, error)
	ExitStatus() (int, error)
	Stdout() (io.Reader, error)
	Stderr() (io.Reader, error)
}

type DefCommand struct {
	Ctx        context.Context
	command    util.ExecCommand
	exitStatus int
	stdout     io.Reader
	stderr     io.Reader
	loaded     bool
	Timeout    int
	err        error
}

func NewDefCommand(ctx context.Context, command interface{}, system *System, config util.Config) (Command, error) {
	switch cmd := command.(type) {
	case string:
		return newDefCommand(ctx, cmd, system, config), nil
	case []string:
		return newDefExecCommand(ctx, cmd, system, config), nil
	default:
		return nil, fmt.Errorf("command type must be either string or []string")
	}
}

func newDefCommand(ctx context.Context, command string, system *System, config util.Config) Command {
	return &DefCommand{
		Ctx:     ctx,
		command: util.ExecCommand{CmdStr: command},
		Timeout: config.TimeOutMilliSeconds(),
	}
}

func newDefExecCommand(ctx context.Context, command []string, system *System, config util.Config) Command {
	return &DefCommand{
		Ctx:     ctx,
		command: util.ExecCommand{CmdSlice: command},
		Timeout: config.TimeOutMilliSeconds(),
	}
}

func (c *DefCommand) setup() error {
	if c.loaded {
		return c.err
	}
	c.loaded = true

	var cmd *util.Command
	if c.command.CmdStr != "" {
		cmd = commandWrapper(c.command.CmdStr)
	} else {
		cmd = util.NewCommand(c.command.CmdSlice[0], c.command.CmdSlice[1:]...)
	}
	err := runCommand(cmd, c.Timeout)

	// We don't care about ExitError since it's covered by status
	if _, ok := err.(*exec.ExitError); !ok {
		c.err = err
	}
	c.exitStatus = cmd.Status
	stdoutB := cmd.Stdout.Bytes()
	stderrB := cmd.Stderr.Bytes()
	id := c.Ctx.Value("id")
	logBytes(stdoutB, fmt.Sprintf("[Command][%s][stdout] ", id))
	logBytes(stderrB, fmt.Sprintf("[Command][%s][stderr] ", id))
	c.stdout = bytes.NewReader(stdoutB)
	c.stderr = bytes.NewReader(stderrB)

	return c.err
}

func (c *DefCommand) Command() util.ExecCommand {
	return c.command
}

func (c *DefCommand) ExitStatus() (int, error) {
	err := c.setup()

	return c.exitStatus, err
}

func (c *DefCommand) Stdout() (io.Reader, error) {
	err := c.setup()

	return c.stdout, err
}

func (c *DefCommand) Stderr() (io.Reader, error) {
	err := c.setup()

	return c.stderr, err
}

// Stub out
func (c *DefCommand) Exists() (bool, error) {
	return false, nil
}

func runCommand(cmd *util.Command, timeout int) error {
	c1 := make(chan bool, 1)
	e1 := make(chan error, 1)
	timeoutD := time.Duration(timeout) * time.Millisecond
	go func() {
		err := cmd.Run()
		if err != nil {
			e1 <- err
		}
		c1 <- true
	}()
	select {
	case <-c1:
		return nil
	case err := <-e1:
		return err
	case <-time.After(timeoutD):
		return fmt.Errorf("Command execution timed out (%s)", timeoutD)
	}
}
