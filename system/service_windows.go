//go:build windows
// +build windows

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
	cmd := util.NewCommandForWindowsPowershell("Get-Service", "-Name", s.service)
	cmd.Run()
	if strings.Contains(cmd.Stderr.String(), "Cannot find any service with service name") {
		return false, nil
	}
	return true, cmd.Err
}

func (s *ServiceWindows) Enabled() (bool, error) {
	cmd := util.NewCommandForWindowsPowershell(fmt.Sprintf("$(Get-Service -Name %s).StartType", s.service))
	cmd.Run()
	if strings.Contains(cmd.Stdout.String(), "Automatic") {
		return true, cmd.Err
	}
	return false, cmd.Err
}

func (s *ServiceWindows) Running() (bool, error) {
	cmd := util.NewCommandForWindowsPowershell(fmt.Sprintf("$(Get-Service -Name %s).Status", s.service))
	cmd.Run()
	if strings.Contains(cmd.Stdout.String(), "Running") {
		return true, cmd.Err
	}
	return false, cmd.Err
}

func (s *ServiceWindows) RunLevels() ([]string, error) {
	return nil, nil
}
