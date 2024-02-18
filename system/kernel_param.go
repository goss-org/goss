package system

import (
	"context"
	"fmt"

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
	value string
}

func NewDefKernelParam(_ context.Context, key interface{}, system *System, config util.Config) (KernelParam, error) {
	strKey, ok := key.(string)
	if !ok {
		return nil, fmt.Errorf("key must be of type string")
	}
	return newDefKernelParam(nil, strKey, system, config), nil
}

func newDefKernelParam(_ context.Context, key string, system *System, config util.Config) KernelParam {
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
