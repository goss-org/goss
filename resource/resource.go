package resource

import "github.com/aelsabbahy/goss/system"

type Resource interface {
	Validate(*system.System) []TestResult
}
