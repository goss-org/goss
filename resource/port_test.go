package resource

import (
	"context"
	"strings"
	"testing"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
	"gopkg.in/yaml.v3"
)

// fakeSysPort is a minimal system.Port implementation used to drive NewPort
// without inspecting the real network stack.
type fakeSysPort struct {
	ips []string
}

func (f *fakeSysPort) Port() string             { return "tcp:9999" }
func (f *fakeSysPort) Exists() (bool, error)    { return false, nil }
func (f *fakeSysPort) Listening() (bool, error) { return false, nil }
func (f *fakeSysPort) IP() ([]string, error)    { return f.ips, nil }

// TestNewPortLeavesIPUnsetWhenEmpty pins the fix for `goss add port` writing
// `ip: []` into generated gossfiles for a port nothing is listening on. An
// empty list asserts nothing, so NewPort must leave IP nil (matching the
// field's yaml:"ip,omitempty" tag) rather than assigning the empty slice.
func TestNewPortLeavesIPUnsetWhenEmpty(t *testing.T) {
	p, err := NewPort(&fakeSysPort{ips: nil}, util.Config{})
	if err != nil {
		t.Fatalf("NewPort returned error: %v", err)
	}
	if p.IP != nil {
		t.Errorf("NewPort().IP = %#v, want nil", p.IP)
	}

	out, err := yaml.Marshal(p)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if strings.Contains(string(out), "ip:") {
		t.Errorf("marshalled yaml contains an ip key, want it omitted:\n%s", out)
	}
}

// TestNewPortKeepsIPWhenPresent makes sure the empty-list guard does not drop
// real values.
func TestNewPortKeepsIPWhenPresent(t *testing.T) {
	p, err := NewPort(&fakeSysPort{ips: []string{"0.0.0.0"}}, util.Config{})
	if err != nil {
		t.Fatalf("NewPort returned error: %v", err)
	}
	out, err := yaml.Marshal(p)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if !strings.Contains(string(out), "0.0.0.0") {
		t.Errorf("marshalled yaml is missing the ip value:\n%s", out)
	}
}

// port.ip used a plain != nil check, so `ip: []` asserted nothing and passed
// silently. It must now warn like file.contents does.
func TestPortEmptyIPWarns(t *testing.T) {
	sys := &system.System{
		NewPort: func(context.Context, string, *system.System, util.Config) system.Port {
			return &fakeSysPort{ips: []string{"0.0.0.0"}}
		},
	}

	p := &Port{id: "tcp:9999", Listening: false, IP: []any{}}
	out := captureStderr(t, func() { p.Validate(sys) })

	if !strings.Contains(out, "WARNING:") || !strings.Contains(out, "port.ip") {
		t.Errorf("Validate with empty ip stderr = %q, want a WARNING naming port.ip", out)
	}
}
