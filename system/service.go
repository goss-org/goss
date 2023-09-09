package system

import (
	"context"
	"strings"
)

type Service interface {
	Service() string
	Exists(context.Context) (bool, error)
	Enabled(context.Context) (bool, error)
	Running(context.Context) (bool, error)
	RunLevels(context.Context) ([]string, error)
}

func invalidService(s string) bool {
	if strings.ContainsRune(s, '/') {
		return true
	}
	return false
}
