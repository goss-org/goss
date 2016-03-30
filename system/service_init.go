package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aelsabbahy/goss/util"
)

type ServiceInit struct {
	service string
	alpine  bool
}

func NewServiceInit(service string, system *System, config util.Config) Service {
	return &ServiceInit{service: service}
}

func NewAlpineServiceInit(service string, system *System, config util.Config) Service {
	return &ServiceInit{service: service, alpine: true}
}

func (s *ServiceInit) Service() string {
	return s.service
}

func (s *ServiceInit) Exists() (bool, error) {
	if _, err := os.Stat(fmt.Sprintf("/etc/init.d/%s", s.service)); err == nil {
		return true, err
	}
	return false, nil
}

func (s *ServiceInit) Enabled() (bool, error) {
	if s.alpine {
		return alpineInitServiceEnabled(s.service, "sysinit")
	} else {
		return initServiceEnabled(s.service, 3)
	}
}

func (s *ServiceInit) Running() (bool, error) {
	cmd := util.NewCommand("service", s.service, "status")
	cmd.Run()
	if cmd.Status == 0 {
		return true, cmd.Err
	}
	return false, nil
}

func initServiceEnabled(service string, level int) (bool, error) {
	matches, err := filepath.Glob(fmt.Sprintf("/etc/rc%d.d/S[0-9][0-9]%s", level, service))
	if err == nil && matches != nil {
		return true, nil
	}
	return false, err
}

func alpineInitServiceEnabled(service string, level string) (bool, error) {
	matches, err := filepath.Glob(fmt.Sprintf("/etc/runlevels/%s/%s", level, service))
	if err == nil && matches != nil {
		return true, nil
	}
	return false, err
}
