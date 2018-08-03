package system

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/aelsabbahy/goss/util"
)

type ServiceInit struct {
	service  string
	alpine   bool
	runlevel string
}

func NewServiceInit(service string, system *System, config util.Config) Service {
	return &ServiceInit{service: service}
}

func NewAlpineServiceInit(service string, system *System, config util.Config) Service {
	runlevel := config.RunLevel
	if runlevel == "" {
		typ := reflect.TypeOf(config)
		f, _ := typ.FieldByName("RunLevel")
		runlevel = f.Tag.Get("default")
	}
	return &ServiceInit{service: service, alpine: true, runlevel: runlevel}
}

func (s *ServiceInit) Service() string {
	return s.service
}

func (s *ServiceInit) Exists() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	if _, err := os.Stat(fmt.Sprintf("/etc/init.d/%s", s.service)); err == nil {
		return true, err
	}
	return false, nil
}

func (s *ServiceInit) Enabled() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	if s.alpine {
		return alpineInitServiceEnabled(s.service, s.runlevel)
	} else {
		return initServiceEnabled(s.service, 3)
	}
}

func (s *ServiceInit) Running() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
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
