package system

import (
	"bytes"
	"io"
	"os/exec"
	"strconv"

	"github.com/aelsabbahy/goss/util"
)

type Command interface {
	Command() string
	Exists() (interface{}, error)
	ExitStatus() (interface{}, error)
	Stdout() (io.Reader, error)
	Stderr() (io.Reader, error)
}

type DefCommand struct {
	command    string
	exitStatus string
	stdout     io.Reader
	stderr     io.Reader
	loaded     bool
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

	cmd := util.NewCommand("sh", "-c", c.command)
	err := cmd.Run()

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
