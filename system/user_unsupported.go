//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package system

import "errors"

var errUnsupportedOS = errors.New("unsupported operating system")

func (u *DefUser) Shell() (string, error) {
	return "", errUnsupportedOS
}
