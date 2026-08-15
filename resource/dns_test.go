package resource

import (
	"strings"
	"testing"

	"github.com/goss-org/goss/util"
	"gopkg.in/yaml.v3"
)

// fakeSysDNS is a minimal system.DNS implementation used to drive NewDNS
// without doing a real lookup.
type fakeSysDNS struct {
	addrs []string
}

func (f *fakeSysDNS) Host() string              { return "no-such-host-xyz.invalid" }
func (f *fakeSysDNS) Addrs() ([]string, error)  { return f.addrs, nil }
func (f *fakeSysDNS) Resolvable() (bool, error) { return false, nil }
func (f *fakeSysDNS) Exists() (bool, error)     { return false, nil }
func (f *fakeSysDNS) Server() string            { return "" }
func (f *fakeSysDNS) Qtype() string             { return "" }

// TestNewDNSLeavesAddrsUnsetWhenEmpty pins the fix for `goss add dns` writing
// `addrs: []` into generated gossfiles for an unresolvable host. An empty list
// asserts nothing, so NewDNS must leave Addrs nil (matching the field's
// yaml:"addrs,omitempty" tag) rather than assigning the empty slice.
func TestNewDNSLeavesAddrsUnsetWhenEmpty(t *testing.T) {
	d, err := NewDNS(&fakeSysDNS{addrs: nil}, util.Config{})
	if err != nil {
		t.Fatalf("NewDNS returned error: %v", err)
	}
	if d.Addrs != nil {
		t.Errorf("NewDNS().Addrs = %#v, want nil", d.Addrs)
	}

	out, err := yaml.Marshal(d)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if strings.Contains(string(out), "addrs:") {
		t.Errorf("marshalled yaml contains an addrs key, want it omitted:\n%s", out)
	}
}

// TestNewDNSKeepsAddrsWhenPresent makes sure the empty-list guard does not drop
// real values.
func TestNewDNSKeepsAddrsWhenPresent(t *testing.T) {
	d, err := NewDNS(&fakeSysDNS{addrs: []string{"127.0.0.1"}}, util.Config{})
	if err != nil {
		t.Fatalf("NewDNS returned error: %v", err)
	}
	out, err := yaml.Marshal(d)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(out), "127.0.0.1") {
		t.Errorf("marshalled yaml is missing the addrs value:\n%s", out)
	}
}
