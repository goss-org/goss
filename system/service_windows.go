package system

import (
	"github.com/aelsabbahy/goss/util"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

// ServiceWindows is used to query windows service manager.
type ServiceWindows struct {
	service string
}

func NewServiceWindows(service string, system *System, config util.Config) Service {
	return &ServiceWindows{service: service}
}

func (s *ServiceWindows) Service() string {
	return s.service
}

func (s *ServiceWindows) Exists() (bool, error) {
	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}
	return s.exists(m)
}

func (s *ServiceWindows) exists(m *mgr.Mgr) (bool, error) {
	svcs, err := m.ListServices()
	if err != nil {
		return false, err
	}
	for _, svc := range svcs {
		if svc == s.service {
			return true, nil
		}
	}
	return false, nil
}

func (s *ServiceWindows) Enabled() (bool, error) {
	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}

	// If service does not exist, it's not enabled.
	ex, err := s.exists(m)
	if !ex || err != nil {
		return false, err
	}

	// Open service and check whether it's enabled (i.e. whether it starts automatically).
	svc, err := m.OpenService(s.service)
	if err != nil {
		return false, err
	}
	cfg, err := svc.Config()
	if err != nil {
		return false, err
	}
	return cfg.StartType != mgr.StartDisabled, nil
}

func (s *ServiceWindows) Running() (bool, error) {
	m, err := mgr.Connect()
	if err != nil {
		return false, err
	}

	// If a service does not exist, it's not running.
	ex, err := s.exists(m)
	if !ex || err != nil {
		return false, err
	}

	// Open service and check whether windows reports it as running.
	svc, err := m.OpenService(s.service)
	if err != nil {
		return false, err
	}
	q, err := svc.Query()
	if err != nil {
		return false, err
	}
	return q.State == windows.SERVICE_RUNNING, nil
}
