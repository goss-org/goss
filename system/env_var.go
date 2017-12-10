package system

import (
	"errors"
	"github.com/aelsabbahy/goss/util"
	"os"
)

type EnvVar interface {
	Key() string
	Exists() (bool, error)
	Value() (string, error)
}

type DefEnvVar struct {
	key   string
	value string
}

func NewDefEnvVar(key string, system *System, config util.Config) EnvVar {
	return &DefEnvVar{
		key: key,
	}
}

func (k *DefEnvVar) ID() string {
	return k.key
}

func (k *DefEnvVar) Key() string {
	return k.key
}

func (k *DefEnvVar) Exists() (bool, error) {
	if _, err := k.Value(); err != nil {
		return false, nil
	}
	return true, nil
}

func (k *DefEnvVar) Value() (string, error) {
	value, exists := os.LookupEnv(k.key)
	if !exists {
		return value, errors.New("environmental variable does not exist")
	}
	return value, nil
}
