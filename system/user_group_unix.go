//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package system

import (
	"bufio"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func groupsForUser(user string, pgid int, grp io.Reader) ([]string, error) {
	s := bufio.NewScanner(grp)
	out := []string{}

	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}

		text := s.Text()
		if text == "" {
			continue
		}

		// see: man 5 group
		//  group_name:password:GID:user_list
		// Name:Pass:Gid:List
		//  root:x:0:root
		//  adm:x:4:root,adm,daemon
		parts := strings.Split(text, ":")
		if len(parts) != 4 {
			continue
		}

		gid, err := strconv.Atoi(parts[2])
		if err == nil {
			if gid == pgid {
				out = append(out, parts[0])
				continue
			}
		}

		for _, g := range strings.Split(parts[3], ",") {
			if g == user {
				out = append(out, parts[0])
				continue
			}
		}
	}

	sort.Strings(out)

	return out, nil
}

func (u *DefUser) Groups() ([]string, error) {
	grp, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer grp.Close()

	pgid, err := u.GID()
	if err != nil {
		return nil, err
	}

	return groupsForUser(u.username, pgid, grp)
}
