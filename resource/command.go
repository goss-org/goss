package resource

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Command struct {
	Title      string   `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta     `json:"meta,omitempty" yaml:"meta,omitempty"`
	Command    string   `json:"-" yaml:"-"`
	ExitStatus matcher  `json:"exit-status" yaml:"exit-status"`
	Stdout     []string `json:"stdout" yaml:"stdout"`
	Stderr     []string `json:"stderr" yaml:"stderr"`
	Timeout    int      `json:"timeout" yaml:"timeout"`
}

func (c *Command) ID() string      { return c.Command }
func (c *Command) SetID(id string) { c.Command = id }

func (c *Command) GetTitle() string { return c.Title }
func (c *Command) GetMeta() meta    { return c.Meta }

func (c *Command) Validate(sys *system.System) []TestResult {
	if c.Timeout == 0 {
		c.Timeout = 10000
	}
	sysCommand := sys.NewCommand(c.Command, sys, util.Config{Timeout: c.Timeout})

	var results []TestResult

	cExitStatus := deprecateAtoI(c.ExitStatus, fmt.Sprintf("%s: command.exit-status", c.Command))
	results = append(results, ValidateValue(c, "exit-status", cExitStatus, sysCommand.ExitStatus))

	if len(c.Stdout) > 0 {
		results = append(results, ValidateContains(c, "stdout", c.Stdout, sysCommand.Stdout))
	}
	if len(c.Stderr) > 0 {
		results = append(results, ValidateContains(c, "stderr", c.Stderr, sysCommand.Stderr))
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
		Timeout:    config.Timeout,
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
