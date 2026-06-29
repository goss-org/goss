//go:build windows
// +build windows

package system

import "errors"

var errNotImplemented = errors.New("Not implemented")

func getUsage(mountpoint string) (int, error) {
	return 0, errNotImplemented
}
