package system

import (
	"context"

	"github.com/achanda/go-sysctl"
	"github.com/goss-org/goss/util"
)

type KernelParam interface {
	Key() string
	Exists() (bool, error)
	Value() (string, error)
}

type DefKernelParam struct {
	key   string
}

func NewDefKernelParam(_ context.Context, key string, system *System, config util.Config) KernelParam {
	return &DefKernelParam{
		key: key,
	}
}

func (k *DefKernelParam) ID() string {
	return k.key
}

func (k *DefKernelParam) Key() string {
	return k.key
}

func (k *DefKernelParam) Exists() (bool, error) {
	if _, err := k.Value(); err != nil {
		return false, nil
	}
	return true, nil
}

func (k *DefKernelParam) Value() (string, error) {
	return sysctl.Get(k.key)
}
