package system

import (
	"context"
	"net"

	"github.com/goss-org/goss/util"
)

type Interface interface {
	Name() string
	Exists(context.Context) (bool, error)
	Addrs(context.Context) ([]string, error)
	MTU(context.Context) (int, error)
}

type DefInterface struct {
	name   string
	loaded bool
	exists bool
	iface  *net.Interface
	err    error
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
	return nil
}

func (i *DefInterface) ID() string {
	return i.name
}

func (i *DefInterface) Name() string {
	return i.name
}

func (i *DefInterface) Exists(ctx context.Context) (bool, error) {
	if err := i.setup(); err != nil {
		return false, nil
	}

	return i.exists, nil
}

func (i *DefInterface) Addrs(ctx context.Context) ([]string, error) {
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

func (i *DefInterface) MTU(ctx context.Context) (int, error) {
	if err := i.setup(); err != nil {
		return 0, err
	}

	return i.iface.MTU, nil
}
