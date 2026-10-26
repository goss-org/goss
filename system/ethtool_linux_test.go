//go:build linux

package system

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// realEthtoolSettings is a trimmed but otherwise verbatim `ethtool eth0` from a
// 10G NIC, kept so the parser is exercised against the tool's actual layout
// (tab indented, values that contain spaces and brackets).
const realEthtoolSettings = "Settings for eth0:\n" +
	"\tSupported ports: [ TP ]\n" +
	"\tSupported link modes:   1000baseT/Full\n" +
	"\t                        10000baseT/Full\n" +
	"\tSupported pause frame use: No\n" +
	"\tSupports auto-negotiation: Yes\n" +
	"\tSpeed: 10000Mb/s\n" +
	"\tDuplex: Full\n" +
	"\tAuto-negotiation: on\n" +
	"\tPort: Twisted Pair\n" +
	"\tPHYAD: 0\n" +
	"\tTransceiver: internal\n" +
	"\tMDI-X: off (auto)\n" +
	"\tCurrent message level: 0x00000007 (7)\n" +
	"\tLink detected: yes\n"

// realEthtoolFeatures is a trimmed `ethtool -k eth0`.
const realEthtoolFeatures = "Features for eth0:\n" +
	"\trx-checksumming: on\n" +
	"\ttx-checksumming: on\n" +
	"\t    tx-checksum-ipv4: on\n" +
	"\tscatter-gather: on\n" +
	"\ttcp-segmentation-offload: on\n" +
	"\tgeneric-receive-offload: on\n" +
	"\tlarge-receive-offload: off [fixed]\n"

func TestParseEthtoolFields(t *testing.T) {
	cases := []struct {
		name string
		key  string
		want string
	}{
		{"speed", "Speed", "10000Mb/s"},
		{"duplex", "Duplex", "Full"},
		{"auto-negotiation", "Auto-negotiation", "on"},
		{"value with spaces and brackets", "Supported ports", "[ TP ]"},
		{"value with a parenthesised number", "Current message level", "0x00000007 (7)"},
		{"last line without a trailing newline", "Link detected", "yes"},
	}

	fields := parseEthtoolFields(realEthtoolSettings)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := fields[c.key]; got != c.want {
				t.Errorf("fields[%q] = %q, want %q", c.key, got, c.want)
			}
		})
	}
}

func TestParseEthtoolFeatureFields(t *testing.T) {
	fields := parseEthtoolFields(realEthtoolFeatures)

	if got := fields["rx-checksumming"]; got != "on" {
		t.Errorf("rx-checksumming = %q, want %q", got, "on")
	}
	if got := fields["tx-checksumming"]; got != "on" {
		t.Errorf("tx-checksumming = %q, want %q", got, "on")
	}
	if got := fields["large-receive-offload"]; got != "off [fixed]" {
		t.Errorf("large-receive-offload = %q, want %q", got, "off [fixed]")
	}
}

func TestParseEthtoolFieldsIgnoresLinesWithoutAKey(t *testing.T) {
	fields := parseEthtoolFields("\n\tSpeed: 1000Mb/s\nnot a field\n: no key\n")

	if len(fields) != 1 {
		t.Errorf("parsed %d fields %v, want only Speed", len(fields), fields)
	}
	if got := fields["Speed"]; got != "1000Mb/s" {
		t.Errorf("Speed = %q, want %q", got, "1000Mb/s")
	}
}

// A value ethtool never printed is an error naming the interface, so a
// baseline over many interfaces stays diagnosable.
func TestDefEthtoolUnreportedFieldNamesTheInterface(t *testing.T) {
	e := &defEthtool{name: "eth9", loaded: true, exists: true, settings: map[string]string{}}

	_, err := e.Speed()
	if err == nil {
		t.Fatal("Speed() = nil error, want one reporting the missing value")
	}
	if !strings.Contains(err.Error(), "eth9") || !strings.Contains(err.Error(), "Speed") {
		t.Errorf("error = %q, want it to name eth9 and Speed", err)
	}
}

// An interface that is simply not present is a failed assertion, not a runtime
// error: Exists() reports false and the attribute getters explain themselves.
func TestDefEthtoolMissingInterface(t *testing.T) {
	e := &defEthtool{name: "definitely-not-an-interface", loaded: true, exists: false}

	exists, err := e.Exists()
	if err != nil {
		t.Fatalf("Exists() = %v, want no error", err)
	}
	if exists {
		t.Error("Exists() = true, want false")
	}

	if _, err := e.RxChecksum(); err == nil {
		t.Fatal("RxChecksum() = nil error, want one reporting the missing interface")
	} else if !strings.Contains(err.Error(), "definitely-not-an-interface") {
		t.Errorf("error = %q, want it to name the interface", err)
	}
}

// Every getter surfaces a missing ethtool binary, so the reason is visible
// instead of a bogus zero value.
func TestDefEthtoolWithoutBinary(t *testing.T) {
	e := &defEthtool{name: "lo", loaded: true, err: ErrEthtoolNotFound, settings: map[string]string{}}

	for name, got := range map[string]func() (string, error){
		"Speed":      e.Speed,
		"Duplex":     e.Duplex,
		"RxChecksum": e.RxChecksum,
		"TxChecksum": e.TxChecksum,
		"AdaptiveRx": e.AdaptiveRx,
		"AdaptiveTx": e.AdaptiveTx,
		"RxUsecs":    e.RxUsecs,
		"TxUsecs":    e.TxUsecs,
	} {
		if _, err := got(); !strings.Contains(err.Error(), ErrEthtoolNotFound.Error()) {
			t.Errorf("%s() error = %v, want it to wrap %v", name, err, ErrEthtoolNotFound)
		}
	}
}

func TestInterfaceNamesIncludesLoopback(t *testing.T) {
	names, err := InterfaceNames()
	if err != nil {
		t.Fatalf("InterfaceNames() returned error: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("InterfaceNames() returned nothing, want at least one interface")
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("InterfaceNames() = %v, want sorted output", names)
			break
		}
	}
}

// stubEthtool puts a fake ethtool on the PATH and points the process at it, so
// Driver can be exercised without depending on the test host's NICs.
func stubEthtool(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ethtool"), []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("writing the ethtool stub: %v", err)
	}
	t.Setenv("PATH", dir)
}

// Driver takes the driver out of "ethtool -i" and caches it, so a baseline that
// only excludes by driver does not pay for a lookup per attribute.
func TestDefEthtoolDriver(t *testing.T) {
	stubEthtool(t, "if [ \"$1\" = \"-i\" ]; then\n"+
		"  echo 'driver: ixgbe'\n"+
		"  echo 'bus-info: 0000:01:00.0'\n"+
		"fi\n")

	e := &defEthtool{name: "lo", loaded: true, exists: true}

	driver, err := e.Driver()
	if err != nil {
		t.Fatalf("Driver() returned error: %v", err)
	}
	if driver != "ixgbe" {
		t.Errorf("Driver() = %q, want %q", driver, "ixgbe")
	}

	// The second call must be served from the cache: the stub now fails.
	stubEthtool(t, "exit 1\n")
	if driver, err := e.Driver(); err != nil || driver != "ixgbe" {
		t.Errorf("Driver() = %q, %v on the second call, want the cached %q", driver, err, "ixgbe")
	}
}

// An interface ethtool has no driver information for fails with a message naming
// the interface, so a baseline over many interfaces stays diagnosable.
func TestDefEthtoolDriverWithoutDriverInfo(t *testing.T) {
	stubEthtool(t, "if [ \"$1\" = \"-i\" ]; then\n"+
		"  echo 'bus-info: 0000:01:00.0'\n"+
		"fi\n")

	e := &defEthtool{name: "lo", loaded: true, exists: true}

	_, err := e.Driver()
	if err == nil {
		t.Fatal("Driver() = nil error, want one reporting the missing driver")
	}
	if !strings.Contains(err.Error(), "lo") || !strings.Contains(err.Error(), "driver") {
		t.Errorf("error = %q, want it to name lo and driver", err)
	}
}

// A missing ethtool binary surfaces through Driver too.
func TestDefEthtoolDriverWithoutBinary(t *testing.T) {
	e := &defEthtool{name: "lo", loaded: true, err: ErrEthtoolNotFound}

	if _, err := e.Driver(); err == nil || !strings.Contains(err.Error(), ErrEthtoolNotFound.Error()) {
		t.Errorf("Driver() error = %v, want it to wrap %v", err, ErrEthtoolNotFound)
	}
}

// InterfaceDrivers leaves out the interfaces whose driver it could not read, so
// that a driver pattern can never match an unknown driver by accident.
func TestInterfaceDriversOmitsUnknownDrivers(t *testing.T) {
	drivers, err := InterfaceDrivers()
	if err != nil {
		t.Skipf("this host cannot report drivers: %v", err)
	}

	for name, driver := range drivers {
		if driver == "" {
			t.Errorf("InterfaceDrivers()[%q] = %q, want such interfaces to be omitted", name, driver)
		}
	}
}

// realEthtoolCoalesce is a trimmed but otherwise verbatim `ethtool -c eth0`.
const realEthtoolCoalesce = "Coalesce parameters for eth0:\n" +
	"Adaptive RX: off  TX: off\n" +
	"stats-block-usecs: 0\n" +
	"sample-interval: 0\n" +
	"pkt-rate-low: 0\n" +
	"pkt-rate-high: 0\n" +
	"\n" +
	"rx-usecs: 3\n" +
	"rx-frames: 0\n" +
	"rx-usecs-irq: 0\n" +
	"rx-frames-irq: 0\n" +
	"\n" +
	"tx-usecs: 3\n" +
	"tx-frames: 0\n" +
	"tx-usecs-irq: 0\n" +
	"tx-frames-irq: 0\n\n"

// Adaptive coalescing is reported on a single line holding both values, so the
// parser splits it out; the usecs delays come from ordinary lines.
func TestParseEthtoolCoalesce(t *testing.T) {
	fields := parseEthtoolCoalesce(realEthtoolCoalesce)

	cases := []struct {
		key  string
		want string
	}{
		{"adaptive_rx", "off"},
		{"adaptive_tx", "off"},
		{"rx-usecs", "3"},
		{"tx-usecs", "3"},
		{"rx-usecs-irq", "0"},
		{"rx-frames", "0"},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			if got := fields[c.key]; got != c.want {
				t.Errorf("fields[%q] = %q, want %q", c.key, got, c.want)
			}
		})
	}
}

func TestParseEthtoolCoalesceWithAdaptiveEnabled(t *testing.T) {
	fields := parseEthtoolCoalesce("Coalesce parameters for eth0:\n" +
		"Adaptive RX: on  TX: off\n" +
		"rx-usecs: 20\n")

	if got := fields["adaptive_rx"]; got != "on" {
		t.Errorf("adaptive_rx = %q, want %q", got, "on")
	}
	if got := fields["adaptive_tx"]; got != "off" {
		t.Errorf("adaptive_tx = %q, want %q", got, "off")
	}
	if got := fields["rx-usecs"]; got != "20" {
		t.Errorf("rx-usecs = %q, want %q", got, "20")
	}
}

// A driver that reports no adaptive line simply has no adaptive values, rather
// than empty ones that an assertion could match by accident.
func TestParseEthtoolCoalesceWithoutAdaptiveLine(t *testing.T) {
	fields := parseEthtoolCoalesce("Coalesce parameters for eth0:\n" + "rx-usecs: 3\n")

	if _, ok := fields["adaptive_rx"]; ok {
		t.Errorf("adaptive_rx = %q, want the key to be absent", fields["adaptive_rx"])
	}
	if _, ok := fields["adaptive_tx"]; ok {
		t.Errorf("adaptive_tx = %q, want the key to be absent", fields["adaptive_tx"])
	}
}

// The four coalescing getters read "ethtool -c" once and cache it, so a rule
// asserting several of them costs a single process spawn.
func TestDefEthtoolCoalescingGetters(t *testing.T) {
	stubEthtool(t, "if [ \"$1\" = \"-c\" ]; then\n"+
		"  echo 'Coalesce parameters for eth0:'\n"+
		"  echo 'Adaptive RX: off  TX: off'\n"+
		"  echo 'rx-usecs: 3'\n"+
		"  echo 'tx-usecs: 3'\n"+
		"fi\n")

	e := &defEthtool{name: "lo", loaded: true, exists: true}

	cases := []struct {
		name string
		want string
		get  func() (string, error)
	}{
		{"AdaptiveRx", "off", e.AdaptiveRx},
		{"AdaptiveTx", "off", e.AdaptiveTx},
		{"RxUsecs", "3", e.RxUsecs},
		{"TxUsecs", "3", e.TxUsecs},
	}
	for _, c := range cases {
		got, err := c.get()
		if err != nil {
			t.Fatalf("%s() returned error: %v", c.name, err)
		}
		if got != c.want {
			t.Errorf("%s() = %q, want %q", c.name, got, c.want)
		}
	}

	// The settings are served from the cache: the stub now fails.
	stubEthtool(t, "exit 1\n")
	if got, err := e.TxUsecs(); err != nil || got != "3" {
		t.Errorf("TxUsecs() = %q, %v after the stub was replaced, want the cached %q", got, err, "3")
	}
}

// An interface whose coalescing settings ethtool refuses to report fails with
// the command's own message instead of a missing-value error.
func TestDefEthtoolCoalescingUnsupported(t *testing.T) {
	stubEthtool(t, "echo 'Cannot get device coalesce settings: Operation not supported' >&2\n"+
		"exit 1\n")

	e := &defEthtool{name: "lo", loaded: true, exists: true}

	_, err := e.AdaptiveRx()
	if err == nil {
		t.Fatal("AdaptiveRx() = nil error, want the ethtool failure")
	}
	if !strings.Contains(err.Error(), "Operation not supported") {
		t.Errorf("error = %q, want it to carry ethtool's message", err)
	}
}

// realEthtoolRing is a verbatim `ethtool -g eth0`: the same four labels appear
// twice, once as the pre-set maximums and once as the current settings.
const realEthtoolRing = "Ring parameters for eth0:\n" +
	"Pre-set maximums:\n" +
	"RX:\t\t4096\n" +
	"RX Mini:\t\t0\n" +
	"RX Jumbo:\t\t0\n" +
	"TX:\t\t4096\n" +
	"Current hardware settings:\n" +
	"RX:\t\t512\n" +
	"RX Mini:\t\t0\n" +
	"RX Jumbo:\t\t0\n" +
	"TX:\t\t512\n\n"

// ringFields reads the current sizes and the pre-set maximums apart, so a rule
// can assert the effective size, the ceiling, or both.
func TestParseEthtoolRing(t *testing.T) {
	fields := parseEthtoolRing(realEthtoolRing)

	cases := map[string]string{
		"rx_ring":           "512",
		"tx_ring":           "512",
		"rx_mini_ring":      "0",
		"rx_jumbo_ring":     "0",
		"rx_ring_max":       "4096",
		"tx_ring_max":       "4096",
		"rx_mini_ring_max":  "0",
		"rx_jumbo_ring_max": "0",
	}
	if len(fields) != len(cases) {
		t.Errorf("parseEthtoolRing() returned %d fields, want %d: %v", len(fields), len(cases), fields)
	}
	for key, want := range cases {
		if got := fields[key]; got != want {
			t.Errorf("fields[%q] = %q, want %q", key, got, want)
		}
	}
}

// Labels goss has no attribute for are dropped rather than invented.
func TestParseEthtoolRingIgnoresUnknownLabels(t *testing.T) {
	fields := parseEthtoolRing("Ring parameters for eth0:\n" +
		"Current hardware settings:\n" +
		"RX:\t\t512\n" +
		"RX-BUFs:\t\t0\n" +
		"TX:\t\t512\n")

	if _, ok := fields["rx_bufs"]; ok {
		t.Errorf("unknown label was mapped to an attribute: %v", fields)
	}
	if fields["rx_ring"] != "512" || fields["tx_ring"] != "512" {
		t.Errorf("parseEthtoolRing() = %v, want the known labels kept", fields)
	}
}

// Without a section header the labels would be ambiguous, so the maximums are
// only recorded under their own header.
func TestParseEthtoolRingWithoutMaximumsSection(t *testing.T) {
	fields := parseEthtoolRing("Ring parameters for eth0:\n" +
		"Current hardware settings:\n" +
		"RX:\t\t512\n" +
		"TX:\t\t512\n")

	if _, ok := fields["rx_ring_max"]; ok {
		t.Errorf("maximum recorded without a maximums section: %v", fields)
	}
	if fields["rx_ring"] != "512" {
		t.Errorf("fields = %v, want rx_ring 512", fields)
	}
}

// The ring attributes are read once and cached, like the other ethtool reads.
func TestDefEthtoolRingGetters(t *testing.T) {
	stubEthtool(t, "if [ \"$1\" = \"-g\" ]; then\n"+
		"  printf '%s\\n' 'Ring parameters for eth0:' 'Pre-set maximums:' 'RX:\t\t4096' 'TX:\t\t4096' 'Current hardware settings:' 'RX:\t\t512' 'TX:\t\t512'\n"+
		"fi\n")

	e := &defEthtool{name: "lo", loaded: true, exists: true}

	cases := []struct {
		name string
		want string
		get  func() (string, error)
	}{
		{"RingRx", "512", e.RingRx},
		{"RingTx", "512", e.RingTx},
		{"RingRxMax", "4096", e.RingRxMax},
		{"RingTxMax", "4096", e.RingTxMax},
	}
	for _, c := range cases {
		got, err := c.get()
		if err != nil {
			t.Fatalf("%s() returned error: %v", c.name, err)
		}
		if got != c.want {
			t.Errorf("%s() = %q, want %q", c.name, got, c.want)
		}
	}

	// The sizes are served from the cache: the stub now fails.
	stubEthtool(t, "exit 1\n")
	if got, err := e.RingRx(); err != nil || got != "512" {
		t.Errorf("RingRx() = %q, %v after the stub was replaced, want the cached %q", got, err, "512")
	}
}

// A label the driver does not report fails with a message naming the interface
// rather than comparing against an empty value.
func TestDefEthtoolRingUnsupported(t *testing.T) {
	stubEthtool(t, "echo 'Cannot get device ring settings: Operation not supported' >&2\n"+
		"exit 1\n")

	e := &defEthtool{name: "lo", loaded: true, exists: true}

	_, err := e.RingRx()
	if err == nil {
		t.Fatal("RingRx() = nil error, want the ethtool failure")
	}
	if !strings.Contains(err.Error(), "Operation not supported") {
		t.Errorf("error = %q, want it to carry ethtool's message", err)
	}
}

// realEthtoolPrivFlags is a verbatim `ethtool --show-priv-flags eth0`: a header
// line, then flags whose names are printed with the values column aligned.
const realEthtoolPrivFlags = "Private flags for eth0:\n" +
	"rx-gro-hw        : off\n" +
	"tx-nocache-copy  : off\n" +
	"loopback         : off\n" +
	"tls-hw-tx-offload: on\n"

// Private flags are reported as sorted "name: value" pairs, and the header line
// is not mistaken for a flag with an empty value.
func TestParsePrivFlags(t *testing.T) {
	got := parsePrivFlags(realEthtoolPrivFlags)
	want := []string{"loopback: off", "rx-gro-hw: off", "tls-hw-tx-offload: on", "tx-nocache-copy: off"}

	if len(got) != len(want) {
		t.Fatalf("parsePrivFlags() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parsePrivFlags()[%d] = %q, want %q (got %v)", i, got[i], want[i], got)
		}
	}
}

// Private flags are read once and cached, like the other ethtool reads.
func TestDefEthtoolPrivFlags(t *testing.T) {
	stubEthtool(t, "if [ \"$1\" = \"--show-priv-flags\" ]; then\n"+
		"  echo 'Private flags for eth0:'\n"+
		"  echo 'loopback         : off'\n"+
		"  echo 'rx-gro-hw        : off'\n"+
		"fi\n")

	e := &defEthtool{name: "lo", loaded: true, exists: true}

	flags, err := e.PrivFlags()
	if err != nil {
		t.Fatalf("PrivFlags() returned error: %v", err)
	}
	if len(flags) != 2 || flags[0] != "loopback: off" || flags[1] != "rx-gro-hw: off" {
		t.Errorf("PrivFlags() = %v, want the sorted flags", flags)
	}

	// The flags are served from the cache: the stub now fails.
	stubEthtool(t, "exit 1\n")
	if flags, err := e.PrivFlags(); err != nil || len(flags) != 2 {
		t.Errorf("PrivFlags() = %v, %v after the stub was replaced, want the cached flags", flags, err)
	}
}

// A driver that exposes no private flags fails with ethtool's own message.
func TestDefEthtoolPrivFlagsUnsupported(t *testing.T) {
	stubEthtool(t, "echo 'Cannot get device priv flags: Operation not supported' >&2\n"+
		"exit 1\n")

	e := &defEthtool{name: "lo", loaded: true, exists: true}

	if _, err := e.PrivFlags(); err == nil || !strings.Contains(err.Error(), "Operation not supported") {
		t.Errorf("PrivFlags() error = %v, want it to carry ethtool's message", err)
	}
}

// The driver lookup reads the sysfs symlink when the kernel exposes one, which
// is the same answer "ethtool -i" gives but without a process per interface.
func TestSysfsDriver(t *testing.T) {
	root := t.TempDir()
	old := sysfsNetPath
	sysfsNetPath = root
	t.Cleanup(func() { sysfsNetPath = old })

	device := filepath.Join(root, "eth0", "device")
	if err := os.MkdirAll(device, 0o755); err != nil {
		t.Fatalf("creating the fixture: %v", err)
	}
	if err := os.Symlink("../../../../bus/pci/drivers/ixgbe", filepath.Join(device, "driver")); err != nil {
		t.Fatalf("creating the driver symlink: %v", err)
	}

	if driver, ok := sysfsDriver("eth0"); !ok || driver != "ixgbe" {
		t.Errorf("sysfsDriver(eth0) = %q, %v, want %q, true", driver, ok, "ixgbe")
	}

	// An interface with no device behind it, and one that does not exist at
	// all, both leave the answer to ethtool.
	if err := os.MkdirAll(filepath.Join(root, "veth0"), 0o755); err != nil {
		t.Fatalf("creating the fixture: %v", err)
	}
	for _, name := range []string{"veth0", "eth9"} {
		if driver, ok := sysfsDriver(name); ok {
			t.Errorf("sysfsDriver(%s) = %q, true, want no driver", name, driver)
		}
	}
}

// InterfaceDrivers is answered from sysfs alone when every interface has a
// device behind it: no ethtool binary is needed, and none is spawned.
func TestInterfaceDriversFromSysfsWithoutEthtool(t *testing.T) {
	names, err := InterfaceNames()
	if err != nil {
		t.Fatalf("InterfaceNames() returned error: %v", err)
	}

	root := t.TempDir()
	old := sysfsNetPath
	sysfsNetPath = root
	t.Cleanup(func() { sysfsNetPath = old })

	for _, name := range names {
		device := filepath.Join(root, name, "device")
		if err := os.MkdirAll(device, 0o755); err != nil {
			t.Fatalf("creating the fixture for %s: %v", name, err)
		}
		if err := os.Symlink("/sys/bus/pci/drivers/ixgbe", filepath.Join(device, "driver")); err != nil {
			t.Fatalf("creating the driver symlink for %s: %v", name, err)
		}
	}

	// An empty PATH means there is no ethtool to fall back to, so any driver in
	// the result can only have come from sysfs.
	t.Setenv("PATH", t.TempDir())

	drivers, err := InterfaceDrivers()
	if err != nil {
		t.Fatalf("InterfaceDrivers() returned error: %v", err)
	}
	if len(drivers) != len(names) {
		t.Fatalf("InterfaceDrivers() = %v, want an entry for each of %v", drivers, names)
	}
	for _, name := range names {
		if drivers[name] != "ixgbe" {
			t.Errorf("InterfaceDrivers()[%q] = %q, want %q", name, drivers[name], "ixgbe")
		}
	}
}
