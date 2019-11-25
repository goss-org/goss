package system

import (
	"fmt"
	"strings"

	"github.com/aelsabbahy/goss/util"
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

func (s *ServiceSystemd) Exists() (bool, error) {
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
		if e, err := sysv.Exists(); e && err == nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *ServiceSystemd) Enabled() (bool, error) {
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
		if en, err := sysv.Enabled(); en && err == nil {
			return true, nil
		}
	}
	return false, nil
}

func (s *ServiceSystemd) Running() (bool, error) {
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
		if r, err := sysv.Running(); r && err == nil {
			return true, nil
		}
	}
	return false, nil
}
