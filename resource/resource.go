package resource

import "github.com/aelsabbahy/goss/system"

type Resource interface {
	Validate(*system.System) []TestResult
	SetID(string)
}

type IDer interface {
	ID() string
}
