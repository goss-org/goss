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

func NewServiceUpstart(service string, system *System, config util.Config) Service {
	return &ServiceUpstart{service: service}
}

func (s *ServiceUpstart) Service() string {
	return s.service
}

func (s *ServiceUpstart) Exists(ctx context.Context) (bool, error) {
	// upstart
	if _, err := os.Stat(fmt.Sprintf("/etc/init/%s.conf", s.service)); err == nil {
		return true, nil
	}
	// Fallback on sysv
	sysv := &ServiceInit{service: s.service}
	if e, err := sysv.Exists(ctx); e && err == nil {
		return true, nil
	}
	return false, nil
}

func (s *ServiceUpstart) Enabled(ctx context.Context) (bool, error) {
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
	if en, err := sysv.Enabled(ctx); en && err == nil {
		return true, nil
	}
	return false, nil
}

func (s *ServiceUpstart) Running(ctx context.Context) (bool, error) {
	cmd := util.NewCommand("service", s.service, "status")
	cmd.Run()
	out := cmd.Stdout.String()
	if cmd.Status == 0 && (strings.Contains(out, "running") || strings.Contains(out, "online")) {
		return true, cmd.Err
	}
	return false, nil
}
func (s *ServiceUpstart) RunLevels(ctx context.Context) ([]string, error) {
	sysv := &ServiceInit{service: s.service}
	return sysv.RunLevels(ctx)
}
