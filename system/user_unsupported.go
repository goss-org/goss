//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package system

import (
	"fmt"
)

func (u *DefUser) Shell() (string, error) {
	return "", fmt.Errorf("unsupported operating system")
}
