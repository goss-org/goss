package system

import (
	"strings"
	"testing"

	"github.com/goss-org/GOnetstat"
)

func TestValidatePortNetwork(t *testing.T) {
	t.Parallel()
	for _, port := range []string{"tcp:22", "tcp6:22", "udp:53", "udp6:53"} {
		if err := validatePortNetwork(port); err != nil {
			t.Errorf("validatePortNetwork(%q) = %v, want nil", port, err)
		}
	}
	for _, port := range []string{"tcp4:22", "udp4:53", "tpc:22", "TCP:22", ":22"} {
		err := validatePortNetwork(port)
		if err == nil {
			t.Errorf("validatePortNetwork(%q) = nil, want an error", port)
			continue
		}
		if !strings.Contains(err.Error(), "tcp, tcp6, udp, udp6") {
			t.Errorf("validatePortNetwork(%q) error %q does not list the valid networks", port, err)
		}
	}
}

// A port with an unknown network can never be found by GetPorts. It must
// report an error rather than "not listening", or `listening: false` passes
// without checking anything (goss-org/goss#1006).
func TestDefPortUnknownNetworkErrors(t *testing.T) {
	t.Parallel()
	listener := []GOnetstat.Process{{Ip: "0.0.0.0", Port: 80, State: "LISTEN"}}
	p := &DefPort{
		port:     "tcp4:80",
		sysPorts: map[string][]GOnetstat.Process{"tcp:80": listener},
		err:      validatePortNetwork("tcp4:80"),
	}

	if listening, err := p.Listening(); err == nil {
		t.Errorf("Listening() = %v, nil; want an error for network tcp4", listening)
	}
	if exists, err := p.Exists(); err == nil {
		t.Errorf("Exists() = %v, nil; want an error for network tcp4", exists)
	}
	if ips, err := p.IP(); err == nil {
		t.Errorf("IP() = %v, nil; want an error for network tcp4", ips)
	}
}

func TestDefPortKnownNetworkUnchanged(t *testing.T) {
	t.Parallel()
	listener := []GOnetstat.Process{{Ip: "0.0.0.0", Port: 80, State: "LISTEN"}}
	ports := map[string][]GOnetstat.Process{"tcp:80": listener}

	open := &DefPort{port: "tcp:80", sysPorts: ports, err: validatePortNetwork("tcp:80")}
	if listening, err := open.Listening(); err != nil || !listening {
		t.Errorf("tcp:80 Listening() = %v, %v; want true, nil", listening, err)
	}
	if ips, err := open.IP(); err != nil || len(ips) != 1 || ips[0] != "0.0.0.0" {
		t.Errorf("tcp:80 IP() = %v, %v; want [0.0.0.0], nil", ips, err)
	}

	closed := &DefPort{port: "tcp6:80", sysPorts: ports, err: validatePortNetwork("tcp6:80")}
	if listening, err := closed.Listening(); err != nil || listening {
		t.Errorf("tcp6:80 Listening() = %v, %v; want false, nil", listening, err)
	}
}

// A bare port number still defaults to tcp.
func TestNormalizePortDefaultsToTCP(t *testing.T) {
	t.Parallel()
	if got := normalizePort("22"); got != "tcp:22" {
		t.Errorf("normalizePort(\"22\") = %q, want \"tcp:22\"", got)
	}
	if err := validatePortNetwork(normalizePort("22")); err != nil {
		t.Errorf("validatePortNetwork(normalizePort(\"22\")) = %v, want nil", err)
	}
}
