package system

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"github.com/aelsabbahy/GOnetstat"
	// This needs a better name
	"github.com/aelsabbahy/go-ps"
	util2 "github.com/aelsabbahy/goss/util"
	"github.com/urfave/cli"
)

type Resource interface {
	Exists() (bool, error)
}

type System struct {
	NewPackage     func(string, *System, util2.Config) Package
	NewFile        func(string, *System, util2.Config) File
	NewAddr        func(string, *System, util2.Config) Addr
	NewPort        func(string, *System, util2.Config) Port
	NewService     func(string, *System, util2.Config) Service
	NewUser        func(string, *System, util2.Config) User
	NewGroup       func(string, *System, util2.Config) Group
	NewCommand     func(string, *System, util2.Config) Command
	NewDNS         func(string, *System, util2.Config) DNS
	NewProcess     func(string, *System, util2.Config) Process
	NewGossfile    func(string, *System, util2.Config) Gossfile
	NewKernelParam func(string, *System, util2.Config) KernelParam
	NewMount       func(string, *System, util2.Config) Mount
	NewInterface   func(string, *System, util2.Config) Interface
	NewHTTP        func(string, *System, util2.Config) HTTP
	ports          map[string][]GOnetstat.Process
	portsOnce      sync.Once
	procMap        map[string][]ps.Process
	procOnce       sync.Once
}

func (s *System) Ports() map[string][]GOnetstat.Process {
	s.portsOnce.Do(func() {
		s.ports = GetPorts(false)
	})
	return s.ports
}

func (s *System) ProcMap() map[string][]ps.Process {
	s.procOnce.Do(func() {
		s.procMap = GetProcs()
	})
	return s.procMap
}

func New(c *cli.Context) *System {
	sys := &System{
		NewFile:        NewDefFile,
		NewAddr:        NewDefAddr,
		NewPort:        NewDefPort,
		NewUser:        NewDefUser,
		NewGroup:       NewDefGroup,
		NewCommand:     NewDefCommand,
		NewDNS:         NewDefDNS,
		NewProcess:     NewDefProcess,
		NewGossfile:    NewDefGossfile,
		NewKernelParam: NewDefKernelParam,
		NewMount:       NewDefMount,
		NewInterface:   NewDefInterface,
		NewHTTP:        NewDefHTTP,
	}
	sys.detectService()
	sys.detectPackage(c)
	return sys
}

// detectPackage adds the correct package creation function to a System struct
func (sys *System) detectPackage(c *cli.Context) {
	p := c.GlobalString("package")
	if p != "deb" && p != "apk" && p != "pacman" && p != "rpm" {
		p = DetectPackageManager()
	}
	switch p {
	case "deb":
		sys.NewPackage = NewDebPackage
	case "apk":
		sys.NewPackage = NewAlpinePackage
	case "pacman":
		sys.NewPackage = NewPacmanPackage
	default:
		sys.NewPackage = NewRpmPackage
	}
}

// detectService adds the correct service creation function to a System struct
func (sys *System) detectService() {
	switch DetectService() {
	case "windows":
		sys.NewService = NewServiceWindows
	case "upstart":
		sys.NewService = NewServiceUpstart
	case "systemd":
		sys.NewService = NewServiceSystemd
	case "alpineinit":
		sys.NewService = NewAlpineServiceInit
	default:
		sys.NewService = NewServiceInit
	}
}

// DetectPackageManager attempts to detect whether or not the system is using
// "deb", "rpm", "apk", or "pacman" package managers. It first attempts to
// detect the distro. If that fails, it falls back to finding package manager
// executables. If that fails, it returns the empty string.
func DetectPackageManager() string {
	switch DetectDistro() {
	case "ubuntu":
		return "deb"
	case "redhat":
		return "rpm"
	case "alpine":
		return "apk"
	case "arch":
		return "pacman"
	case "debian":
		return "deb"
	}
	for _, manager := range []string{"deb", "rpm", "apk", "pacman"} {
		if HasCommand(manager) {
			return manager
		}
	}
	return ""
}

// DetectService attempts to detect what kind of service management the system
// is using, "systemd", "upstart", "alpineinit", or "init". It looks for systemctl
// command to detect systemd, and falls back on DetectDistro otherwise. If it can't
// decide, it returns "init".
func DetectService() string {
	if runtime.GOOS == "windows" {
		return "windows"
	}
	if HasCommand("systemctl") {
		return "systemd"
	}
	// Centos Docker container doesn't run systemd, so we detect it or use init.
	switch DetectDistro() {
	case "ubuntu":
		return "upstart"
	case "alpine":
		return "alpineinit"
	case "arch":
		return "systemd"
	}
	return "init"
}

// DetectDistro attempts to detect which Linux distribution this computer is
// using. One of "ubuntu", "redhat" (including Centos), "alpine", "arch", or
// "debian". If it can't decide, it returns an empty string.
func DetectDistro() string {
	if b, e := ioutil.ReadFile("/etc/lsb-release"); e == nil && bytes.Contains(b, []byte("Ubuntu")) {
		return "ubuntu"
	} else if isRedhat() {
		return "redhat"
	} else if _, err := os.Stat("/etc/alpine-release"); err == nil {
		return "alpine"
	} else if _, err := os.Stat("/etc/arch-release"); err == nil {
		return "arch"
	} else if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "debian"
	}
	return ""
}

// HasCommand returns whether or not an executable by this name is on the PATH.
func HasCommand(cmd string) bool {
	if _, err := exec.LookPath(cmd); err == nil {
		return true
	}
	return false
}

func isRedhat() bool {
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return true
	} else if _, err := os.Stat("/etc/system-release"); err == nil {
		return true
	}
	return false
}
