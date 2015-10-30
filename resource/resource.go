package resource

import (
	"path/filepath"

	"github.com/aelsabbahy/goss/system"
)

type Resource interface {
	Validate(*system.System) []TestResult
	SetID(string)
}

type IDer interface {
	ID() string
}

func contains(a []string, s string) bool {
	for _, e := range a {
		if m, _ := filepath.Match(e, s); m {
			return true
		}
	}
	return false
}
