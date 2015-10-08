package system

import (
	"bytes"
	"io"
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
}

func NewDefCommand(command string, system *System) Command {
	return &DefCommand{command: command}
}

func (c *DefCommand) setup() {
	if c.loaded {
		return
	}
	c.loaded = true

	cmd := util.NewCommand("sh", "-c", c.command)
	cmd.Run()

	c.exitStatus = strconv.Itoa(cmd.Status)
	c.stdout = bytes.NewReader(cmd.Stdout.Bytes())
	c.stderr = bytes.NewReader(cmd.Stderr.Bytes())
}

func (c *DefCommand) Command() string {
	return c.command
}

func (c *DefCommand) ExitStatus() (interface{}, error) {
	c.setup()

	return c.exitStatus, nil
}

func (c *DefCommand) Stdout() (io.Reader, error) {
	c.setup()

	return c.stdout, nil
}

func (c *DefCommand) Stderr() (io.Reader, error) {
	c.setup()

	return c.stderr, nil
}

// Stub out
func (c *DefCommand) Exists() (interface{}, error) {
	return false, nil
}
