package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/goss-org/goss/util"
)

type ServiceWindows struct {
	service string
}

func NewServiceWindows(_ context.Context, service string, system *System, config util.Config) Service {
	return &ServiceWindows{
		service: service,
	}
}

func (s *ServiceWindows) Service() string {
	return s.service
}

func (s *ServiceWindows) Exists() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	cmd := util.NewCommand("systemctltest", "-q", "list-unit-files", "--type=service")
	cmd.Run()
	if strings.Contains(cmd.Stdout.String(), fmt.Sprintf("%s.service", s.service)) {
		return true, cmd.Err
	}
	return false, nil
}

func (s *ServiceWindows) Enabled() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	cmd := util.NewCommand("systemctltest", "-q", "is-enabled", s.service)
	cmd.Run()
	if cmd.Status == 0 {
		return true, cmd.Err
	}
	return false, nil
}

func (s *ServiceWindows) Running() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	cmd := util.NewCommand("systemctltest", "-q", "is-active", s.service)
	cmd.Run()
	if cmd.Status == 0 {
		return true, cmd.Err
	}
	return false, nil
}

func (s *ServiceWindows) RunLevels() ([]string, error) {
	return nil, nil
}

// test