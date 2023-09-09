package system

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/goss-org/goss/util"
)

type Command interface {
	Command() string
	Exists(context.Context) (bool, error)
	ExitStatus(context.Context) (int, error)
	Stdout(context.Context) (io.Reader, error)
	Stderr(context.Context) (io.Reader, error)
}

type DefCommand struct {
	command    string
	exitStatus int
	stdout     io.Reader
	stderr     io.Reader
	loaded     bool
	Timeout    int
	err        error
}

func NewDefCommand(command string, system *System, config util.Config) Command {
	return &DefCommand{
		command: command,
		Timeout: config.TimeOutMilliSeconds(),
	}
}

func (c *DefCommand) setup() error {
	if c.loaded {
		return c.err
	}
	c.loaded = true

	cmd := commandWrapper(c.command)
	err := runCommand(cmd, c.Timeout)

	// We don't care about ExitError since it's covered by status
	if _, ok := err.(*exec.ExitError); !ok {
		c.err = err
	}
	c.exitStatus = cmd.Status
	stdoutB := cmd.Stdout.Bytes()
	stderrB := cmd.Stderr.Bytes()
	logBytes(stdoutB, fmt.Sprintf("Command: %s: stdout: ", "validate-pytorch"))
	logBytes(stderrB, fmt.Sprintf("Command: %s: stderr: ", "validate-pytorch"))
	c.stdout = bytes.NewReader(stdoutB)
	c.stderr = bytes.NewReader(stderrB)

	return c.err
}

func logBytes(b []byte, prefix string) {
	if len(b) == 0 {
		return
	}
	lines := bytes.Split(b, []byte("\n"))
	for _, l := range lines {
		log.Printf("[DEBUG] %s %s", prefix, l)
	}
}

func (c *DefCommand) Command() string {
	return c.command
}

func (c *DefCommand) ExitStatus(ctx context.Context) (int, error) {
	err := c.setup()

	return c.exitStatus, err
}

func (c *DefCommand) Stdout(ctx context.Context) (io.Reader, error) {
	err := c.setup()

	return c.stdout, err
}

func (c *DefCommand) Stderr(ctx context.Context) (io.Reader, error) {
	err := c.setup()

	return c.stderr, err
}

// Stub out
func (c *DefCommand) Exists(ctx context.Context) (bool, error) {
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
