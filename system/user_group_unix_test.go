//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package system

import (
	"strings"
	"testing"
)

func TestGroupsForUser(t *testing.T) {
	grp := `badline
testgrp1:*:100:bob,jack,jill
testgrp2:*:101:bob,jack
testgrp3:*:102:jill
testgrp4:*:103:`

	var cases = []struct {
		user   string
		gid    int
		expect []string
	}{
		{"bob", 100, []string{"testgrp1", "testgrp2"}},
		{"jack", 100, []string{"testgrp1", "testgrp2"}},
		{"jill", 103, []string{"testgrp1", "testgrp3", "testgrp4"}},
		{"other", 103, []string{"testgrp4"}},
		{"other", 105, []string{}},
	}

	for _, c := range cases {
		res, err := groupsForUser(c.user, c.gid, strings.NewReader(grp))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != len(c.expect) {
			t.Fatalf("result %#v does not match %#v", res, c.expect)
		}
		for i, e := range c.expect {
			if res[i] != e {
				t.Fatalf("result %#v does not match %#v", res, c.expect)
			}
		}
	}
}
