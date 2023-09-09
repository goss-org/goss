package system

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/goss-org/goss/util"
)

type ServiceInit struct {
	service  string
	alpine   bool
	runlevel string
}

func NewServiceInit(_ context.Context, service string, system *System, config util.Config) Service {
	return &ServiceInit{service: service}
}

func NewAlpineServiceInit(_ context.Context, service string, system *System, config util.Config) Service {
	runlevel := config.RunLevel
	if runlevel == "" {
		runlevel = "sysinit"
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
	var runLevels []string
	var err error
	if s.alpine {
		runLevels, err = alpineServiceRunLevels(s.service)
	} else {
		runLevels, err = initServiceRunLevels(s.service)
	}
	return len(runLevels) != 0, err
}

func (s *ServiceInit) RunLevels() ([]string, error) {
	if invalidService(s.service) {
		return nil, nil
	}
	if s.alpine {
		return alpineServiceRunLevels(s.service)
	} else {
		return initServiceRunLevels(s.service)
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

func initServiceRunLevels(service string) ([]string, error) {
	var runLevels []string
	matches, err := filepath.Glob(fmt.Sprintf("/etc/rc*.d/S[0-9][0-9]%s", service))
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile("/etc/rc([0-9]+).d/")
	for _, m := range matches {
		matches := re.FindStringSubmatch(m)
		if matches != nil {
			runLevels = append(runLevels, matches[1])
		}
	}
	return runLevels, nil
}

func alpineServiceRunLevels(service string) ([]string, error) {
	var runLevels []string
	matches, err := filepath.Glob(fmt.Sprintf("/etc/runlevels/*/%s", service))
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile("/etc/runlevels/([^/]+)")
	for _, m := range matches {
		matches := re.FindStringSubmatch(m)
		if matches != nil {
			runLevels = append(runLevels, matches[1])
		}
	}
	return runLevels, nil
}
