package resource

import (
	"bufio"
	"io"
	"strings"

	"github.com/aelsabbahy/goss/system"
)

type Command struct {
	Command    string   `json:"command"`
	ExitStatus string   `json:"exit-status"`
	Stdout     []string `json:"stdout"`
	Stderr     []string `json:"stderr"`
}

func (c *Command) Validate(sys *system.System) []TestResult {
	syscommand := sys.NewCommand(c.Command, sys)

	var results []TestResult

	results = append(results, ValidateValue(c.Command, "exit-status", c.ExitStatus, syscommand.ExitStatus))
	if len(c.Stdout) > 0 {
		results = append(results, ValidateContains(c.Command, "stdout", c.Stdout, syscommand.Stdout))
	}
	if len(c.Stderr) > 0 {
		results = append(results, ValidateContains(c.Command, "stderr", c.Stderr, syscommand.Stderr))
	}

	return results
}

func NewCommand(sysCommand system.Command) *Command {
	command := sysCommand.Command()
	exitStatus, _ := sysCommand.ExitStatus()
	stdout, _ := sysCommand.Stdout()
	stderr, _ := sysCommand.Stderr()
	return &Command{
		Command:    command,
		ExitStatus: exitStatus.(string),
		Stdout:     readerToSlice(stdout),
		Stderr:     readerToSlice(stderr),
	}
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
