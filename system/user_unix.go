//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package system

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

func (u *DefUser) Shell(ctx context.Context) (string, error) {
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return "", err
	}
	defer passwd.Close()

	lines := bufio.NewReader(passwd)

	for {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}

		fs := strings.Split(string(line), ":")
		if len(fs) != 7 {
			return "", fmt.Errorf("invalid entry in /etc/passwd")
		}

		if fs[0] == u.username {
			return fs[6], nil
		}
	}

	return "", fmt.Errorf("unknown user %s", u.username)
}
