//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris
// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris

package system

import (
	"context"
	"fmt"
)

func (u *DefUser) Shell(ctx context.Context) (string, error) {
	return "", fmt.Errorf("unsupported operating system")
}
