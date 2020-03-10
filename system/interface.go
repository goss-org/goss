package system

import (
	"net"

	"github.com/aelsabbahy/goss/util"

	"github.com/jaypipes/ghw"
)

type Interface interface {
	Name() string
	Exists() (bool, error)
	Addrs() ([]string, error)
	MTU() (int, error)
	MAC() (string, error)
	IsVirtual() (bool, error)
}

type DefInterface struct {
	name     string
	loaded   bool
	exists   bool
	iface    *net.Interface
	ghwIface *ghw.NIC
	err      error
}

func NewDefInterface(name string, systei *System, config util.Config) Interface {
	return &DefInterface{
		name: name,
	}
}

func (i *DefInterface) setup() error {
	if i.loaded {
		return i.err
	}
	i.loaded = true

	iface, err := net.InterfaceByName(i.name)
	if err != nil {
		i.exists = false
		i.err = err
		return i.err
	}
	i.iface = iface
	i.exists = true
	ghwNetwork, err := ghw.Network()
	if err != nil {
		i.exists = false
		i.err = err
		return i.err
	}
	for _, nic := range ghwNetwork.NICs {
		if nic.Name == i.name {
			i.ghwIface = nic
			break
		}
	}
	return nil
}

func (i *DefInterface) ID() string {
	return i.name
}

// MAC to get interface's MAC address
func (i *DefInterface) MAC() (string, error) {
	if err := i.setup(); err != nil {
		return "", err
	}
	return i.ghwIface.MacAddress, nil
}

// IsVirtual check if interface is virtual device
func (i *DefInterface) IsVirtual() (bool, error) {
	if err := i.setup(); err != nil {
		return false, nil
	}
	return i.ghwIface.IsVirtual, nil
}

func (i *DefInterface) Name() string {
	return i.name
}

func (i *DefInterface) Exists() (bool, error) {
	if err := i.setup(); err != nil {
		return false, nil
	}

	return i.exists, nil
}

func (i *DefInterface) Addrs() ([]string, error) {
	if err := i.setup(); err != nil {
		return nil, err
	}

	addrs, err := i.iface.Addrs()
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, addr := range addrs {
		ret = append(ret, addr.String())
	}
	return ret, nil
}

func (i *DefInterface) MTU() (int, error) {
	if err := i.setup(); err != nil {
		return 0, err
	}

	return i.iface.MTU, nil
}
