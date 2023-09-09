package system

import (
	"context"

	"github.com/achanda/go-sysctl"
	"github.com/goss-org/goss/util"
)

type KernelParam interface {
	Key() string
	Exists(context.Context) (bool, error)
	Value(context.Context) (string, error)
}

type DefKernelParam struct {
	key   string
	value string
}

func NewDefKernelParam(key string, system *System, config util.Config) KernelParam {
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

func (k *DefKernelParam) Exists(ctx context.Context) (bool, error) {
	if _, err := k.Value(ctx); err != nil {
		return false, nil
	}
	return true, nil
}

func (k *DefKernelParam) Value(ctx context.Context) (string, error) {
	return sysctl.Get(k.key)
}
