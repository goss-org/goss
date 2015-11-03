package system

import (
	"strings"

	"github.com/coreos/go-systemd/dbus"
)

type ServiceDbus struct {
	service string
	dbus    *dbus.Conn
}

func NewServiceDbus(service string, system *System) Service {
	return &ServiceDbus{
		service: service,
		dbus:    system.Dbus,
	}
}

func (s *ServiceDbus) Service() string {
	return s.service
}

func (s *ServiceDbus) Exists() (interface{}, error) {
	units, err := s.dbus.ListUnits()
	if err != nil {
		return false, err
	}
	for _, u := range units {
		if u.Name == s.service+".service" {
			return true, nil
		}
	}
	return false, err
}

func (s *ServiceDbus) Enabled() (interface{}, error) {
	stateRaw, err := s.dbus.GetUnitProperty(s.service+".service", "UnitFileState")
	if err != nil {
		return false, err
	}
	state := stateRaw.Value.String()
	state = strings.Trim(state, "\"")

	if state == "enabled" {
		return true, nil
	}

	// Fall back on initv
	if en, _ := initServiceEnabled(s.service, 3); en {
		return true, nil
	}

	return false, nil
}

func (s *ServiceDbus) Running() (interface{}, error) {
	stateRaw, err := s.dbus.GetUnitProperty(s.service+".service", "ActiveState")
	if err != nil {
		return false, err
	}
	state := stateRaw.Value.String()
	state = strings.Trim(state, "\"")

	if state == "active" {
		return true, nil
	}

	return false, nil
}
