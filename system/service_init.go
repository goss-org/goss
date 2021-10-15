package system

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/aelsabbahy/goss/util"
)

type ServiceInit struct {
	service string
	alpine  bool
	freebsd bool
}

func NewServiceInit(service string, system *System, config util.Config) Service {
	return &ServiceInit{service: service}
}

func NewAlpineServiceInit(service string, system *System, config util.Config) Service {
	return &ServiceInit{service: service, alpine: true}
}

// NewFreeBSDServiceInit returns ServiceInit structure for FreeBSD.
func NewFreeBSDServiceInit(service string, system *System, config util.Config) Service {
	return &ServiceInit{service: service, freebsd: true}
}

func (s *ServiceInit) Service() string {
	return s.service
}

func (s *ServiceInit) Exists() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	if s.freebsd {
		return freebsdInitServiceExists(s.service)
	} else {
		return initServiceExists(s.service)
	}
}

func (s *ServiceInit) Enabled() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	if s.alpine {
		return alpineInitServiceEnabled(s.service, "sysinit")
	} else if s.freebsd {
		return bsdInitServiceEnabled(s.service, "freebsd")
	} else {
		return initServiceEnabled(s.service, 3)
	}
}

func (s *ServiceInit) Running() (bool, error) {
	if invalidService(s.service) {
		return false, nil
	}
	var cmd *util.Command
	if s.freebsd {
		cmd = util.NewCommand("service", s.service, "onestatus")
	} else {
		cmd = util.NewCommand("service", s.service, "status")
	}
	cmd.Run()
	if cmd.Status == 0 {
		return true, cmd.Err
	}
	return false, nil
}

func initServiceExists(service string) (bool, error) {
	if _, err := os.Stat(fmt.Sprintf("/etc/init.d/%s", service)); err == nil {
		return true, err
	}
	return false, nil
}

func freebsdInitServiceExists(service string) (bool, error) {
	searchDir := [...]string{"/etc/rc.d", "/usr/local/etc/rc.d"}

	for _, dir := range searchDir {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", dir, service)); err == nil {
			return true, err
		}
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

func bsdInitServiceEnabled(service string, opsys string) (bool, error) {
	rcconfs, err := filepath.Glob("/etc/rc.conf*")
	if err != nil {
		return false, err
	}

	for _, rcconf := range rcconfs {
		f, err := os.Open(rcconf)
		if err != nil {
			return false, err
		}
		defer f.Close()

		var r *regexp.Regexp
		switch opsys {
		case "freebsd":
			r = regexp.MustCompile(fmt.Sprintf("^%s_enable=\"?(YES|yes)\"?", service))
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if r.MatchString(scanner.Text()) {
				return true, nil
			}
		}
	}
	return false, err
}
