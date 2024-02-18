package resource

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Command struct {
	Title      string           `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta             `json:"meta,omitempty" yaml:"meta,omitempty"`
	id         string           `json:"-" yaml:"-"`
	Exec       util.ExecCommand `json:"exec,omitempty" yaml:"exec,omitempty"`
	ExitStatus matcher          `json:"exit-status" yaml:"exit-status"`
	Stdout     matcher          `json:"stdout" yaml:"stdout"`
	Stderr     matcher          `json:"stderr" yaml:"stderr"`
	Timeout    int              `json:"timeout" yaml:"timeout"`
	Skip       bool             `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	CommandResourceKey  = "command"
	CommandResourceName = "Command"
)

func init() {
	registerResource(CommandResourceKey, &Command{})
}

func (c *Command) ID() string       { return c.id }
func (c *Command) SetID(id string)  { c.id = id }
func (c *Command) SetSkip()         { c.Skip = true }
func (c *Command) TypeKey() string  { return CommandResourceKey }
func (c *Command) TypeName() string { return CommandResourceName }

func (c *Command) GetTitle() string { return c.Title }
func (c *Command) GetMeta() meta    { return c.Meta }
func (c *Command) GetExec() interface{} {
	if c.Exec.CmdStr != "" {
		return c.Exec.CmdStr
	} else if len(c.Exec.CmdSlice) > 0 {
		return c.Exec.CmdSlice
	} else {
		return util.ExecCommand{CmdStr: c.id}
	}
}

func (c *Command) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), "id", c.ID())
	skip := c.Skip

	if c.Timeout == 0 {
		c.Timeout = 10000
	}

	var results []TestResult
	sysCommand, err := sys.NewCommand(ctx, c.GetExec(), sys, util.Config{Timeout: time.Duration(c.Timeout) * time.Millisecond})
	if err != nil {
		log.Printf("[ERROR] Could not create new command: %v", err)
		startTime := time.Now()
		results = append(
			results,
			TestResult{
				Result:       FAIL,
				ResourceType: "Command",
				ResourceId:   c.id,
				Title:        c.Title,
				Meta:         c.Meta,
				Property:     "type",
				Err:          toValidateError(err),
				StartTime:    startTime,
				EndTime:      startTime,
				Duration:     startTime.Sub(startTime),
			},
		)
	} else {
		cExitStatus := deprecateAtoI(c.ExitStatus, fmt.Sprintf("%s: command.exit-status", c.ID()))
		results = append(results, ValidateValue(c, "exit-status", cExitStatus, sysCommand.ExitStatus, skip))
		if isSet(c.Stdout) {
			results = append(results, ValidateValue(c, "stdout", c.Stdout, sysCommand.Stdout, skip))
		}
		if isSet(c.Stderr) {
			results = append(results, ValidateValue(c, "stderr", c.Stderr, sysCommand.Stderr, skip))
		}
	}
	return results
}

func NewCommand(sysCommand system.Command, config util.Config) (*Command, error) {
	var id string
	if sysCommand.Command().CmdStr != "" {
		id = sysCommand.Command().CmdStr
	} else {
		id = sysCommand.Command().CmdSlice[0]
	}
	exitStatus, err := sysCommand.ExitStatus()
	c := &Command{
		id:         id,
		ExitStatus: exitStatus,
		Stdout:     "",
		Stderr:     "",
		Timeout:    config.TimeOutMilliSeconds(),
	}

	if !contains(config.IgnoreList, "stdout") {
		stdout, _ := sysCommand.Stdout()
		outSlice := readerToSlice(stdout)
		if len(outSlice) != 0 {
			c.Stdout = outSlice
		}
	}
	if !contains(config.IgnoreList, "stderr") {
		stderr, _ := sysCommand.Stderr()
		errSlice := readerToSlice(stderr)
		if len(errSlice) != 0 {
			c.Stderr = errSlice
		}
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
