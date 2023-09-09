package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/goss-org/goss/util"
)

type ServiceSystemd struct {
	service string
	legacy  bool
}

func NewServiceSystemd(service string, system *System, config util.Config) Service {
	return &ServiceSystemd{
		service: service,
	}
}

func NewServiceSystemdLegacy(service string, system *System, config util.Config) Service {
	return &ServiceSystemd{
		service: service,
		legacy:  true,
	}
}

func (s *ServiceSystemd) Service() string {
	return s.service
}

func (s *ServiceSystemd) Exists(ctx context.Context) (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	cmd := util.NewCommand("systemctl", "-q", "list-unit-files", "--type=service")
	cmd.Run()
	if strings.Contains(cmd.Stdout.String(), fmt.Sprintf("%s.service", s.service)) {
		return true, cmd.Err
	}
	if s.legacy {
		// Fallback on sysv
		sysv := &ServiceInit{service: s.service}
		if e, err := sysv.Exists(ctx); e && err == nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *ServiceSystemd) Enabled(ctx context.Context) (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	cmd := util.NewCommand("systemctl", "-q", "is-enabled", s.service)
	cmd.Run()
	if cmd.Status == 0 {
		return true, cmd.Err
	}
	if s.legacy {
		// Fallback on sysv
		sysv := &ServiceInit{service: s.service}
		if en, err := sysv.Enabled(ctx); en && err == nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *ServiceSystemd) Running(ctx context.Context) (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	cmd := util.NewCommand("systemctl", "-q", "is-active", s.service)
	cmd.Run()
	if cmd.Status == 0 {
		return true, cmd.Err
	}
	if s.legacy {
		// Fallback on sysv
		sysv := &ServiceInit{service: s.service}
		if r, err := sysv.Running(ctx); r && err == nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *ServiceSystemd) RunLevels(ctx context.Context) ([]string, error) {
	return nil, nil
}
