package system

import (
	"fmt"
	"path/filepath"

	"github.com/aelsabbahy/goss/util"
)

type ServiceInit struct {
	service string
	enabled bool
	running bool
}

func NewServiceInit(service string, system *System) Service {
	return &ServiceInit{service: service}
}

func (s *ServiceInit) Service() string {
	return s.service
}

func (s *ServiceInit) Enabled() (interface{}, error) {
	matches, err := filepath.Glob(fmt.Sprintf("/etc/rc3.d/S[0-9][0-9]%s", s.service))
	if err != nil || matches == nil {
		return false, err
	}
	return true, nil
}

func (s *ServiceInit) Running() (interface{}, error) {
	cmd := util.NewCommand("service", s.service, "status")
	cmd.Run()
	if cmd.Status == 0 {
		return true, nil
	}
	return false, nil
}
