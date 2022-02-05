//go:build windows
// +build windows

package system

import "errors"

func getUsage(mountpoint string) (int, error) {
	return 0, errors.New("Not implemented")
}
