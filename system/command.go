package system

import (
	"bytes"
	"io"
	"strconv"

	"github.com/aelsabbahy/goss/util"
)

type Command struct {
	command    string
	exitStatus string
	stdout     io.Reader
	stderr     io.Reader
	loaded     bool
}

func NewCommand(command string, system *System) Command {
	return Command{command: command}
}

func (c *Command) setup() {
	if c.loaded {
		return
	}
	c.loaded = true
	//cmd_array := strings.Fields(c.command)

	cmd := util.NewCommand("sh", "-c", c.command)
	cmd.Run()

	c.exitStatus = strconv.Itoa(cmd.Status)
	c.stdout = bytes.NewReader(cmd.Stdout.Bytes())
	c.stderr = bytes.NewReader(cmd.Stderr.Bytes())
}

func (c *Command) Command() string {
	return c.command
}

func (c *Command) ExitStatus() (interface{}, error) {
	c.setup()

	return c.exitStatus, nil
}

func (c *Command) Stdout() (io.Reader, error) {
	c.setup()

	return c.stdout, nil
}

func (c *Command) Stderr() (io.Reader, error) {
	c.setup()

	return c.stderr, nil
}

// Stub out
func (c *Command) Exists() (interface{}, error) {
	return false, nil
}
