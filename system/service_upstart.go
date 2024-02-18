package system

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/goss-org/goss/util"
)

type ServiceUpstart struct {
	service string
}

var upstartEnabled = regexp.MustCompile(`^\s*start on`)
var upstartDisabled = regexp.MustCompile(`^manual`)

func NewServiceUpstart(_ context.Context, service interface{}, system *System, config util.Config) (Service, error) {
	strService, ok := service.(string)
	if !ok {
		return nil, fmt.Errorf("service must be of type string")
	}
	return newServiceUpstart(nil, strService, system, config), nil
}

func newServiceUpstart(_ context.Context, service string, system *System, config util.Config) Service {
	return &ServiceUpstart{service: service}
}

func (s *ServiceUpstart) Service() string {
	return s.service
}

func (s *ServiceUpstart) Exists() (bool, error) {
	// upstart
	if _, err := os.Stat(fmt.Sprintf("/etc/init/%s.conf", s.service)); err == nil {
		return true, nil
	}
	// Fallback on sysv
	sysv := &ServiceInit{service: s.service}
	if e, err := sysv.Exists(); e && err == nil {
		return true, nil
	}
	return false, nil
}

func (s *ServiceUpstart) Enabled() (bool, error) {
	if fh, err := os.Open(fmt.Sprintf("/etc/init/%s.override", s.service)); err == nil {
		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			line := scanner.Text()
			if upstartDisabled.MatchString(line) {
				return false, nil
			}
		}
	}

	// If no /etc/init/<service>.override with `manual` keyword in it has been found
	// Check the contents of the upstart manifest.
	if fh, err := os.Open(fmt.Sprintf("/etc/init/%s.conf", s.service)); err == nil {
		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			line := scanner.Text()
			if upstartEnabled.MatchString(line) {
				return true, nil
			}
		}
	}
	// Fallback on sysv
	sysv := &ServiceInit{service: s.service}
	if en, err := sysv.Enabled(); en && err == nil {
		return true, nil
	}
	return false, nil
}

func (s *ServiceUpstart) Running() (bool, error) {
	cmd := util.NewCommand("service", s.service, "status")
	cmd.Run()
	out := cmd.Stdout.String()
	if cmd.Status == 0 && (strings.Contains(out, "running") || strings.Contains(out, "online")) {
		return true, cmd.Err
	}
	return false, nil
}
func (s *ServiceUpstart) RunLevels() ([]string, error) {
	sysv := &ServiceInit{service: s.service}
	return sysv.RunLevels()
}
