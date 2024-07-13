package system

import "strings"

type Service interface {
	Service() string
	Exists() (bool, error)
	Enabled() (bool, error)
	Running() (bool, error)
	RunLevels() ([]string, error)
}

func invalidService(s string) bool {
	return strings.ContainsRune(s, '/')
}
