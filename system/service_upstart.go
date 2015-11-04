package system

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aelsabbahy/goss/util"
)

type ServiceUpstart struct {
	service string
}

var upstartEnabled = regexp.MustCompile(`^\s*start on`)

func NewServiceUpstart(service string, system *System) Service {
	return &ServiceUpstart{service: service}
}

func (s *ServiceUpstart) Service() string {
	return s.service
}

func (s *ServiceUpstart) Exists() (interface{}, error) {
	// upstart
	if _, err := os.Stat(fmt.Sprintf("/etc/init/%s.conf", s.service)); err == nil {
		return true, err
	}

	// initv
	if _, err := os.Stat(fmt.Sprintf("/etc/init.d/%s", s.service)); err == nil {
		return true, err
	}
	return false, nil
}

func (s *ServiceUpstart) Enabled() (interface{}, error) {
	if fh, err := os.Open(fmt.Sprintf("/etc/init/%s.conf", s.service)); err == nil {
		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			line := scanner.Text()
			if upstartEnabled.MatchString(line) {
				return true, nil
			}
		}
	}

	// Fall back on initv
	if en, _ := initServiceEnabled(s.service, 3); en {
		return true, nil
	}

	return false, nil
}

func (s *ServiceUpstart) Running() (interface{}, error) {
	cmd := util.NewCommand("service", s.service, "status")
	cmd.Run()
	out := cmd.Stdout.String()
	if cmd.Status == 0 && (strings.Contains(out, "running") || strings.Contains(out, "online")) {
		return true, nil
	}
	return false, nil
}
