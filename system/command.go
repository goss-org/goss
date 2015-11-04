package system

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"

	"github.com/aelsabbahy/goss/util"
)

type Command interface {
	Command() string
	Exists() (interface{}, error)
	ExitStatus() (interface{}, error)
	Stdout() (io.Reader, error)
	Stderr() (io.Reader, error)
	SetTimeout(int64)
}

type DefCommand struct {
	command    string
	exitStatus string
	stdout     io.Reader
	stderr     io.Reader
	loaded     bool
	Timeout    int64
	err        error
}

func NewDefCommand(command string, system *System) Command {
	return &DefCommand{command: command}
}

func (c *DefCommand) setup() error {
	if c.loaded {
		return c.err
	}
	c.loaded = true

	timeout := c.Timeout
	if timeout == 0 {
		timeout = (10 * int64(time.Second) / int64(time.Millisecond))
	}
	cmd := util.NewCommand("sh", "-c", c.command)
	err := runCommand(cmd, timeout)

	// We don't care about ExitError since it's covered by status
	if _, ok := err.(*exec.ExitError); !ok {
		c.err = err
	}
	c.exitStatus = strconv.Itoa(cmd.Status)
	c.stdout = bytes.NewReader(cmd.Stdout.Bytes())
	c.stderr = bytes.NewReader(cmd.Stderr.Bytes())

	return c.err
}

func (c *DefCommand) Command() string {
	return c.command
}

func (c *DefCommand) SetTimeout(t int64) {
	c.Timeout = t
}

func (c *DefCommand) ExitStatus() (interface{}, error) {
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
func (c *DefCommand) Exists() (interface{}, error) {
	return false, nil
}

func runCommand(cmd *util.Command, timeout int64) error {
	c1 := make(chan bool, 1)
	e1 := make(chan error, 1)
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
	case <-time.After(time.Millisecond * time.Duration(timeout)):
		return fmt.Errorf("Command execution timed out (%d milliseconds)", timeout)
	}
}
