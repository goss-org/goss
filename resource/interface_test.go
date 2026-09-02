package resource

import (
	"strings"
	"testing"

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
		t.Errorf("marshalled yaml contains an addrs key, want it omitted:\n%s", out)
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
		t.Errorf("marshalled yaml is missing the addrs value:\n%s", out)
	}
}
