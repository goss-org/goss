//go:build !windows

package system

import (
	"context"

	"github.com/goss-org/goss/util"
)

type defRegistry struct {
	key string
}

func NewDefRegistry(_ context.Context, key string, system *System, config util.Config) Registry {
	return &defRegistry{key: key}
}

func (r *defRegistry) Key() string            { return r.key }
func (r *defRegistry) Exists() (bool, error)  { return false, ErrRegistryUnsupported }
func (r *defRegistry) Value() (string, error) { return "", ErrRegistryUnsupported }
func (r *defRegistry) Type() (string, error)  { return "", ErrRegistryUnsupported }
