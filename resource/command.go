package resource

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Command struct {
	Title      string     `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta       `json:"meta,omitempty" yaml:"meta,omitempty"`
	id         string     `json:"-" yaml:"-"`
	Exec       string     `json:"exec,omitempty" yaml:"exec,omitempty"`
	ExitStatus matcher    `json:"exit-status" yaml:"exit-status"`
	Stdout     matcher    `json:"stdout" yaml:"stdout"`
	Stderr     matcher    `json:"stderr" yaml:"stderr"`
	Timeout    int        `json:"timeout" yaml:"timeout"`
	Skip       bool       `json:"skip,omitempty" yaml:"skip,omitempty"`
	RetryCount int        `json:"retry_count,omitempty" yaml:"retry_count,omitempty"`
	RetryDelay RetryDelay `json:"retry_delay,omitempty" yaml:"retry_delay,omitempty"`
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
func (c *Command) GetExec() string {
	if c.Exec != "" {
		return c.Exec
	}
	return c.id
}
func (c *Command) GetRetryCount() int        { return c.RetryCount }
func (c *Command) GetRetryDelay() RetryDelay { return c.RetryDelay }

func (c *Command) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), system.CommandIDKey, c.ID())
	skip := c.Skip

	if c.Timeout == 0 {
		c.Timeout = 10000
	}

	var results []TestResult

	runWithRetry(c.RetryCount, c.RetryDelay, func() bool {
		sysCommand := sys.NewCommand(ctx, c.GetExec(), sys, util.Config{Timeout: time.Duration(c.Timeout) * time.Millisecond})

		results = []TestResult{}
		cExitStatus := deprecateAtoI(c.ExitStatus, fmt.Sprintf("%s: command.exit-status", c.ID()))
		results = append(results, ValidateValue(c, "exit-status", cExitStatus, sysCommand.ExitStatus, skip))
		if isSetWarnEmpty(c.Stdout, fmt.Sprintf("%s: command.stdout", c.ID())) {
			results = append(results, ValidateValue(c, "stdout", c.Stdout, sysCommand.Stdout, skip))
		}
		if isSetWarnEmpty(c.Stderr, fmt.Sprintf("%s: command.stderr", c.ID())) {
			results = append(results, ValidateValue(c, "stderr", c.Stderr, sysCommand.Stderr, skip))
		}

		return allTestsPassed(results)
	})

	return results
}

func allTestsPassed(results []TestResult) bool {
	for _, r := range results {
		if r.Result != SUCCESS {
			return false
		}
	}
	return true
}

func NewCommand(sysCommand system.Command, config util.Config) (*Command, error) {
	command := sysCommand.Command()
	exitStatus, err := sysCommand.ExitStatus()
	c := &Command{
		id:         command,
		ExitStatus: exitStatus,
		Stdout:     "",
		Stderr:     "",
		Timeout:    config.TimeOutMilliSeconds(),
	}

	if !contains(config.IgnoreList, "stdout") {
		stdout, _ := sysCommand.Stdout()
		if out := readerToMatcher(stdout, config.ExactMatch); out != nil {
			c.Stdout = out
		}
	}
	if !contains(config.IgnoreList, "stderr") {
		stderr, _ := sysCommand.Stderr()
		if out := readerToMatcher(stderr, config.ExactMatch); out != nil {
			c.Stderr = out
		}
	}

	return c, err
}

// readerToMatcher turns captured command output into the matcher goss writes
// for it. By default that's a list of trimmed lines, matched as patterns. With
// exactMatch it's the raw output as a single string, which goss validates for
// an exact (whitespace and newline sensitive) match. It returns nil when there
// is nothing to assert so the caller can leave the default empty value in place.
func readerToMatcher(reader io.Reader, exactMatch bool) matcher {
	if exactMatch {
		b, _ := io.ReadAll(reader)
		if len(b) == 0 {
			return nil
		}
		return string(b)
	}
	slice := readerToSlice(reader)
	if len(slice) == 0 {
		return nil
	}
	return slice
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
