package resource

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Command struct {
	Title      string   `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta     `json:"meta,omitempty" yaml:"meta,omitempty"`
	Command    string   `json:"-" yaml:"-"`
	Exec       string   `json:"exec,omitempty" yaml:"exec,omitempty"`
	ExitStatus matcher  `json:"exit-status" yaml:"exit-status"`
	Stdout     []string `json:"stdout" yaml:"stdout"`
	Stderr     []string `json:"stderr" yaml:"stderr"`
	Timeout    int      `json:"timeout" yaml:"timeout"`
	Skip       bool     `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	CommandResourceKey  = "command"
	CommandResourceName = "Command"
)

func init() {
	registerResource(CommandResourceKey, &Command{})
}

func (c *Command) ID() string       { return c.Command }
func (c *Command) SetID(id string)  { c.Command = id }
func (c *Command) SetSkip()         { c.Skip = true }
func (c *Command) TypeKey() string  { return CommandResourceKey }
func (c *Command) TypeName() string { return CommandResourceName }

func (c *Command) GetTitle() string { return c.Title }
func (c *Command) GetMeta() meta    { return c.Meta }
func (c *Command) GetExec() string {
	if c.Exec != "" {
		return c.Exec
	}
	return c.Command
}

func (c *Command) Validate(sys *system.System) []TestResult {
	skip := c.Skip

	if c.Timeout == 0 {
		c.Timeout = 10000
	}

	var results []TestResult
	sysCommand := sys.NewCommand(c.GetExec(), sys, util.Config{Timeout: time.Duration(c.Timeout) * time.Millisecond})

	cExitStatus := deprecateAtoI(c.ExitStatus, fmt.Sprintf("%s: command.exit-status", c.Command))
	results = append(results, ValidateValue(c, "exit-status", cExitStatus, sysCommand.ExitStatus, skip))
	if len(c.Stdout) > 0 {
		results = append(results, ValidateContains(c, "stdout", c.Stdout, sysCommand.Stdout, skip))
	}
	if len(c.Stderr) > 0 {
		results = append(results, ValidateContains(c, "stderr", c.Stderr, sysCommand.Stderr, skip))
	}
	return results
}

func NewCommand(sysCommand system.Command, config util.Config) (*Command, error) {
	command := sysCommand.Command()
	exitStatus, err := sysCommand.ExitStatus()
	c := &Command{
		Command:    command,
		ExitStatus: exitStatus,
		Stdout:     []string{},
		Stderr:     []string{},
		Timeout:    config.TimeOutMilliSeconds(),
	}

	if !contains(config.IgnoreList, "stdout") {
		stdout, _ := sysCommand.Stdout()
		c.Stdout = readerToSlice(stdout)
	}
	if !contains(config.IgnoreList, "stderr") {
		stderr, _ := sysCommand.Stderr()
		c.Stderr = readerToSlice(stderr)
	}

	return c, err
}

func escapePattern(s string) string {
	if strings.HasPrefix(s, "!") || strings.HasPrefix(s, "/") {
		return "\\" + s
	}
	return s
}

func readerToSlice(reader io.Reader) []string {
	scanner := bufio.NewScanner(reader)
	slice := []string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = escapePattern(line)
		if line != "" {
			slice = append(slice, line)
		}
	}

	return slice
}
