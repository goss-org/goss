package resource

import "github.com/aelsabbahy/goss/system"

type Resource interface {
	Validate(*system.System) []TestResult
	SetID(string)
}

type IDer interface {
	ID() string
}

func contains(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}
	return false
}
