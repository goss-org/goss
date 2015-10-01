package system

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/util"
)

type Resource interface{}

type System struct {
	NewPackage  func(string, *System) Package
	NewFile     func(string, *System) *File
	NewAddr     func(string, *System) *Addr
	NewPort     func(string, *System) *Port
	NewService  func(string, *System) Service
	NewUser     func(string, *System) *User
	NewGroup    func(string, *System) *Group
	NewCommand  func(string, *System) *Command
	NewDNS      func(string, *System) *DNS
	NewProcess  func(string, *System) *Process
	NewGossfile func(string, *System) *Gossfile
	Dbus        *dbus.Conn
	ports       map[string]map[string]string
	portsOnce   sync.Once
}

func (s *System) Ports() map[string]map[string]string {
	s.portsOnce.Do(func() {
		s.ports = GetPorts()
	})
	return s.ports
}

func New(c *cli.Context) *System {
	system := &System{
		NewFile:     NewFile,
		NewAddr:     NewAddr,
		NewPort:     NewPort,
		NewUser:     NewUser,
		NewGroup:    NewGroup,
		NewCommand:  NewCommand,
		NewDNS:      NewDNS,
		NewProcess:  NewProcess,
		NewGossfile: NewGossfile,
	}

	if util.IsRunningSystemd() {
		system.NewService = NewServiceDbus
		dbus, err := dbus.New()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		system.Dbus = dbus
	} else {
		system.NewService = NewServiceInit
	}

	switch {
	case isRpm() || c.GlobalString("package") == "rpm":
		system.NewPackage = NewPackageRpm
	case isDeb() || c.GlobalString("package") == "deb":
		system.NewPackage = NewPackageDeb
	default:
		system.NewPackage = NewPackageNull
	}

	return system
}

func isDeb() bool {
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return true
	}

	// See if it has only one of the package managers
	if hasCommand("dpkg") && !hasCommand("rpm") {
		return true
	}

	return false
}

func isRpm() bool {
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return true
	}

	if _, err := os.Stat("/etc/system-release"); err == nil {
		return true
	}

	// See if it has only one of the package managers
	if hasCommand("rpm") && !hasCommand("dpkg") {
		return true
	}
	return false
}

func hasCommand(cmd string) bool {
	if _, err := exec.LookPath(cmd); err == nil {
		return true
	}
	return false
}
