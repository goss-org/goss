package resource

import (
	"strings"
	"testing"

	"github.com/goss-org/goss/util"
	"gopkg.in/yaml.v3"
)

// fakeSysUser is a minimal system.User implementation used to drive NewUser
// without reading the real passwd/group databases.
type fakeSysUser struct {
	groups []string
}

func (f *fakeSysUser) Username() string          { return "nobody" }
func (f *fakeSysUser) Exists() (bool, error)     { return true, nil }
func (f *fakeSysUser) UID() (int, error)         { return 65534, nil }
func (f *fakeSysUser) GID() (int, error)         { return 65534, nil }
func (f *fakeSysUser) Groups() ([]string, error) { return f.groups, nil }
func (f *fakeSysUser) Home() (string, error)     { return "/nonexistent", nil }
func (f *fakeSysUser) Shell() (string, error)    { return "/usr/sbin/nologin", nil }

// TestNewUserLeavesGroupsUnsetWhenEmpty pins the same empty-list problem as the
// file and http generators: a user belonging to no group must not produce
// `groups: []`, which asserts nothing.
func TestNewUserLeavesGroupsUnsetWhenEmpty(t *testing.T) {
	u, err := NewUser(&fakeSysUser{groups: nil}, util.Config{})
	if err != nil {
		t.Fatalf("NewUser returned error: %v", err)
	}
	if u.Groups != nil {
		t.Errorf("NewUser().Groups = %#v, want nil", u.Groups)
	}

	out, err := yaml.Marshal(u)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if strings.Contains(string(out), "groups:") {
		t.Errorf("marshalled yaml contains a groups key, want it omitted:\n%s", out)
	}
}

// TestNewUserKeepsGroupsWhenPresent makes sure the empty-list guard does not
// drop real values.
func TestNewUserKeepsGroupsWhenPresent(t *testing.T) {
	u, err := NewUser(&fakeSysUser{groups: []string{"nogroup"}}, util.Config{})
	if err != nil {
		t.Fatalf("NewUser returned error: %v", err)
	}
	out, err := yaml.Marshal(u)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(out), "nogroup") {
		t.Errorf("marshalled yaml is missing the groups value:\n%s", out)
	}
}
