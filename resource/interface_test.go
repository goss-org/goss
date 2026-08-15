package resource

import (
	"context"
	"strings"
	"testing"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
	"gopkg.in/yaml.v3"
)

// fakeSysInterface is a minimal system.Interface implementation used to drive
// NewInterface without touching real network interfaces.
type fakeSysInterface struct {
	addrs []string
}

func (f *fakeSysInterface) Name() string             { return "eth0" }
func (f *fakeSysInterface) Exists() (bool, error)    { return true, nil }
func (f *fakeSysInterface) Addrs() ([]string, error) { return f.addrs, nil }
func (f *fakeSysInterface) MTU() (int, error)        { return 1500, nil }

// TestNewInterfaceLeavesAddrsUnsetWhenEmpty pins the same empty-list problem as
// the file and http generators: an interface with no assigned addresses must
// not produce `addrs: []`, which asserts nothing.
func TestNewInterfaceLeavesAddrsUnsetWhenEmpty(t *testing.T) {
	i, err := NewInterface(&fakeSysInterface{addrs: nil}, util.Config{})
	if err != nil {
		t.Fatalf("NewInterface returned error: %v", err)
	}
	if i.Addrs != nil {
		t.Errorf("NewInterface().Addrs = %#v, want nil", i.Addrs)
	}

	out, err := yaml.Marshal(i)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if strings.Contains(string(out), "addrs:") {
		t.Errorf("marshalled YAML contains an 'addrs' field, want it omitted:\n%s", out)
	}
}

// TestNewInterfaceKeepsAddrsWhenPresent makes sure the empty-list guard does
// not drop real values.
func TestNewInterfaceKeepsAddrsWhenPresent(t *testing.T) {
	i, err := NewInterface(&fakeSysInterface{addrs: []string{"10.0.0.1/24"}}, util.Config{})
	if err != nil {
		t.Fatalf("NewInterface returned error: %v", err)
	}
	out, err := yaml.Marshal(i)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(out), "10.0.0.1/24") {
		t.Errorf("marshalled YAML is missing the 'addrs' field:\n%s", out)
	}
}

// interface.addrs used a plain != nil check, so `addrs: []` asserted nothing
// and passed silently. It must now warn like file.contents does.
func TestInterfaceEmptyAddrsWarns(t *testing.T) {
	sys := &system.System{
		NewInterface: func(context.Context, string, *system.System, util.Config) system.Interface {
			return &fakeSysInterface{addrs: []string{"10.0.0.1/24"}}
		},
	}

	i := &Interface{id: "eth0", Exists: true, Addrs: []any{}}
	out := captureStderr(t, func() { i.Validate(sys) })

	if !strings.Contains(out, "WARNING:") || !strings.Contains(out, "interface.addrs") {
		t.Errorf("Validate with empty 'addrs' field, stderr = %q, want a WARNING naming interface.addrs", out)
	}
}
