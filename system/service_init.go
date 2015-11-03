package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aelsabbahy/goss/util"
)

type ServiceInit struct {
	service string
}

func NewServiceInit(service string, system *System) Service {
	return &ServiceInit{service: service}
}

func (s *ServiceInit) Service() string {
	return s.service
}

func (s *ServiceInit) Exists() (interface{}, error) {
	if _, err := os.Stat(fmt.Sprintf("/etc/init.d/%s", s.service)); err == nil {
		return true, err
	}
	return false, nil
}

func (s *ServiceInit) Enabled() (interface{}, error) {
	en, err := initServiceEnabled(s.service, 3)
	if en {
		return true, nil
	}
	return false, err
}

func (s *ServiceInit) Running() (interface{}, error) {
	cmd := util.NewCommand("service", s.service, "status")
	cmd.Run()
	if cmd.Status == 0 {
		return true, nil
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
