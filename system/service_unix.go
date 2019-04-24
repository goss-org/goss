// +build !windows

package system

import (
	"github.com/aelsabbahy/goss/util"
)

// ServiceWindows in service_unix.go is a stub counterpart to ServiceWindows in service_windows.go
type ServiceWindows struct{}

func NewServiceWindows(_ string, _ *System, _ util.Config) Service {
	panic("ServiceWindows used on non-windows platform")
}

func (_ *ServiceWindows) Service() string {
	panic("ServiceWindows used on non-windows platform")
}

func (_ *ServiceWindows) Exists() (bool, error) {
	panic("ServiceWindows used on non-windows platform")
}

func (_ *ServiceWindows) Enabled() (bool, error) {
	panic("ServiceWindows used on non-windows platform")
}

func (_ *ServiceWindows) Running() (bool, error) {
	panic("ServiceWindows used on non-windows platform")
}
