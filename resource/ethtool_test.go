package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

// fakeSysEthtool is a minimal system.Ethtool implementation, used to drive the
// ethtool resource without running ethtool(8).
type fakeSysEthtool struct {
	name           string
	driver         string
	speed          string
	duplex         string
	rxChecksum     string
	txChecksum     string
	adaptiveRx     string
	adaptiveTx     string
	rxUsecs        string
	txUsecs        string
	rxRing         string
	txRing         string
	rxMiniRing     string
	rxJumboRing    string
	rxRingMax      string
	txRingMax      string
	rxMiniRingMax  string
	rxJumboRingMax string
	privFlags      []string
	err            error
}

func (f *fakeSysEthtool) Name() string { return f.name }

func (f *fakeSysEthtool) Exists() (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return true, nil
}

func (f *fakeSysEthtool) Speed() (string, error)      { return f.value(f.speed) }
func (f *fakeSysEthtool) Duplex() (string, error)     { return f.value(f.duplex) }
func (f *fakeSysEthtool) RxChecksum() (string, error) { return f.value(f.rxChecksum) }
func (f *fakeSysEthtool) TxChecksum() (string, error) { return f.value(f.txChecksum) }
func (f *fakeSysEthtool) AdaptiveRx() (string, error) { return f.value(f.adaptiveRx) }
func (f *fakeSysEthtool) AdaptiveTx() (string, error) { return f.value(f.adaptiveTx) }
func (f *fakeSysEthtool) RxUsecs() (string, error)    { return f.value(f.rxUsecs) }
func (f *fakeSysEthtool) TxUsecs() (string, error)    { return f.value(f.txUsecs) }
func (f *fakeSysEthtool) RingRx() (string, error)     { return f.value(f.rxRing) }
func (f *fakeSysEthtool) RingTx() (string, error)     { return f.value(f.txRing) }
func (f *fakeSysEthtool) RingRxMini() (string, error) { return f.value(f.rxMiniRing) }
func (f *fakeSysEthtool) RingRxJumbo() (string, error) {
	return f.value(f.rxJumboRing)
}
func (f *fakeSysEthtool) RingRxMax() (string, error)     { return f.value(f.rxRingMax) }
func (f *fakeSysEthtool) RingTxMax() (string, error)     { return f.value(f.txRingMax) }
func (f *fakeSysEthtool) RingRxMiniMax() (string, error) { return f.value(f.rxMiniRingMax) }
func (f *fakeSysEthtool) RingRxJumboMax() (string, error) {
	return f.value(f.rxJumboRingMax)
}
func (f *fakeSysEthtool) Driver() (string, error) { return f.value(f.driver) }

func (f *fakeSysEthtool) PrivFlags() ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.privFlags) == 0 {
		return nil, fmt.Errorf("ethtool did not report any private flags for interface %s", f.name)
	}
	return f.privFlags, nil
}

func (f *fakeSysEthtool) value(v string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if v == "" {
		return "", fmt.Errorf("ethtool did not report a value for interface %s", f.name)
	}
	return v, nil
}

// fakeEthtoolSystem returns a System whose interface list and ethtool values
// are fixed, so a default rule can be expanded without touching the host.
func fakeEthtoolSystem(names []string, ifaces map[string]fakeSysEthtool) *system.System {
	return fakeEthtoolSystemWithDrivers(names, ifaces, nil)
}

// fakeEthtoolSystemWithDrivers fixes the driver reported for each interface as
// well, so exclude_drivers can be exercised without running ethtool(8).
func fakeEthtoolSystemWithDrivers(names []string, ifaces map[string]fakeSysEthtool, drivers map[string]string) *system.System {
	return &system.System{
		NewEthtool: func(_ context.Context, name string, _ *system.System, _ util.Config) system.Ethtool {
			iface, ok := ifaces[name]
			if !ok {
				return &fakeSysEthtool{name: name, err: errors.New("interface does not exist")}
			}
			iface.name = name
			return &iface
		},
		InterfaceNames:   func() ([]string, error) { return names, nil },
		InterfaceDrivers: func() (map[string]string, error) { return drivers, nil },
	}
}

// baseline is the rule set from the documented example: a 10G full-duplex link
// with both checksum offloads enabled.
func baseline() *Ethtool {
	return &Ethtool{
		id:         EthtoolDefaultID,
		Speed:      "10000Mb/s",
		Duplex:     "Full",
		RxChecksum: "on",
		TxChecksum: "on",
	}
}

func goodInterface() fakeSysEthtool {
	return fakeSysEthtool{
		speed:          "10000Mb/s",
		duplex:         "Full",
		rxChecksum:     "on",
		txChecksum:     "on",
		adaptiveRx:     "off",
		adaptiveTx:     "off",
		rxUsecs:        "3",
		txUsecs:        "3",
		rxRing:         "512",
		txRing:         "512",
		rxMiniRing:     "0",
		rxJumboRing:    "0",
		rxRingMax:      "4096",
		txRingMax:      "4096",
		rxMiniRingMax:  "0",
		rxJumboRingMax: "0",
		privFlags:      []string{"loopback: off", "rx-gro-hw: off", "tx-nocache-copy: off"},
	}
}

// coalescingBaseline asserts the coalescing settings on every interface: the
// latency tuning that interrupts are not deferred and not adapted to traffic.
func coalescingBaseline() *Ethtool {
	return &Ethtool{
		id:         EthtoolDefaultID,
		AdaptiveRx: "off",
		AdaptiveTx: "off",
		RxUsecs:    "3",
		TxUsecs:    "3",
	}
}

// ringBaseline asserts the ring buffers every interface has to use: 512
// descriptors each way, which the driver has to be able to grow to 4096.
func ringBaseline() *Ethtool {
	return &Ethtool{
		id:             EthtoolDefaultID,
		RxRing:         "512",
		TxRing:         "512",
		RxMiniRing:     "0",
		RxJumboRing:    "0",
		RxRingMax:      "4096",
		TxRingMax:      "4096",
		RxMiniRingMax:  "0",
		RxJumboRingMax: "0",
	}
}

// privFlagsBaseline asserts two of the driver's private flags. A list has to be
// contained in what the interface reports, so a private flag the kernel adds in
// a later release does not break the rule.
func privFlagsBaseline() *Ethtool {
	return &Ethtool{
		id:        EthtoolDefaultID,
		PrivFlags: []any{"rx-gro-hw: off", "tx-nocache-copy: off"},
	}
}

// resourceIDs returns the sorted set of resource ids seen in results.
func resourceIDs(results []TestResult) map[string]int {
	ids := make(map[string]int)
	for _, r := range results {
		ids[r.ResourceId]++
	}
	return ids
}

// A default entry is expanded onto every interface on the host, and each
// expanded result carries the interface in its id so a failure is traceable.
func TestEthtoolDefaultExpandsOntoEveryInterface(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0", "eth1"}, map[string]fakeSysEthtool{
		"eth0": goodInterface(),
		"eth1": goodInterface(),
	})

	results := baseline().Validate(t.Context(), sys)

	if len(results) != 8 {
		t.Fatalf("got %d results, want 8 (4 attributes x 2 interfaces)", len(results))
	}
	for _, r := range results {
		if r.Result != SUCCESS {
			t.Errorf("result for %s.%s failed: %v", r.ResourceId, r.Property, r.Err)
		}
		if r.ResourceType != EthtoolResourceName {
			t.Errorf("result for %s has resource type %q, want %q", r.ResourceId, r.ResourceType, EthtoolResourceName)
		}
	}

	ids := resourceIDs(results)
	for _, want := range []string{"default: eth0", "default: eth1"} {
		if ids[want] != 4 {
			t.Errorf("resource id %q appeared %d times, want 4 (got ids %v)", want, ids[want], ids)
		}
	}
}

// exclude_interfaces drops matching interfaces; patterns are globs, so a single
// "veth*" entry covers every veth pair on the host.
func TestEthtoolDefaultHonoursExcludeInterfaces(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"docker0", "eth0", "lo", "veth1234", "vethabcd"}, map[string]fakeSysEthtool{
		"eth0": goodInterface(),
	})

	e := baseline()
	e.ExcludeInterfaces = []string{"lo", "docker0", "veth*"}

	results := e.Validate(t.Context(), sys)

	if len(results) != 4 {
		t.Fatalf("got %d results, want 4 (only eth0 is left after exclusions)", len(results))
	}
	ids := resourceIDs(results)
	if ids["default: eth0"] != 4 {
		t.Errorf("expected only 'default: eth0' to be validated, got ids %v", ids)
	}
}

// exclude_drivers drops interfaces by the driver the kernel uses for them, so a
// host can exempt whole classes of interface ("veth", "bridge") without naming
// them one by one.
func TestEthtoolDefaultHonoursExcludeDrivers(t *testing.T) {
	sys := fakeEthtoolSystemWithDrivers(
		[]string{"br-abc", "eth0", "veth1234"},
		map[string]fakeSysEthtool{"eth0": goodInterface()},
		map[string]string{"br-abc": "bridge", "eth0": "ixgbe", "veth1234": "veth"},
	)

	e := baseline()
	e.ExcludeDrivers = []string{"veth", "bridge"}

	results := e.Validate(t.Context(), sys)

	if len(results) != 4 {
		t.Fatalf("got %d results, want 4 (only eth0 is left after driver exclusions)", len(results))
	}
	ids := resourceIDs(results)
	if ids["default: eth0"] != 4 {
		t.Errorf("expected only 'default: eth0' to be validated, got ids %v", ids)
	}
}

// A driver pattern is a glob just like an interface pattern, so "virtio_*"
// covers every virtio NIC on the host.
func TestEthtoolDefaultExcludeDriversMatchesGlob(t *testing.T) {
	sys := fakeEthtoolSystemWithDrivers(
		[]string{"ens3", "ens4", "eth0"},
		map[string]fakeSysEthtool{"eth0": goodInterface()},
		map[string]string{"ens3": "virtio_net", "ens4": "virtio_net", "eth0": "ixgbe"},
	)

	e := baseline()
	e.ExcludeDrivers = []string{"virtio_*"}

	results := e.Validate(t.Context(), sys)

	ids := resourceIDs(results)
	if ids["default: eth0"] != 4 {
		t.Errorf("expected eth0 to still be validated, got ids %v", ids)
	}
	for _, name := range []string{"ens3", "ens4"} {
		if ids["default: "+name] != 0 {
			t.Errorf("%s is driven by virtio_net and must be excluded, got ids %v", name, ids)
		}
	}
}

// An interface whose driver cannot be determined is never excluded by a driver
// pattern, not even a catch-all one; it has to be named in exclude_interfaces.
func TestEthtoolDefaultDoesNotExcludeInterfacesWithUnknownDriver(t *testing.T) {
	sys := fakeEthtoolSystemWithDrivers(
		[]string{"eth0", "lo"},
		map[string]fakeSysEthtool{"eth0": goodInterface(), "lo": {}},
		map[string]string{"eth0": "ixgbe"}, // lo has no driver information at all
	)

	e := baseline()
	e.ExcludeDrivers = []string{"*"}

	results := e.Validate(t.Context(), sys)

	ids := resourceIDs(results)
	if ids["default: eth0"] != 0 {
		t.Errorf("eth0 has a known driver matching *, want it excluded, got ids %v", ids)
	}
	if ids["default: lo"] != 4 {
		t.Errorf("an unknown driver must not be exempted by a driver pattern, got ids %v", ids)
	}
}

// A failure while looking drivers up is reported against the default entry
// rather than dropping interfaces silently.
func TestEthtoolDefaultDriverListError(t *testing.T) {
	sys := &system.System{
		NewEthtool: func(_ context.Context, name string, _ *system.System, _ util.Config) system.Ethtool {
			return &fakeSysEthtool{name: name}
		},
		InterfaceNames:   func() ([]string, error) { return []string{"eth0"}, nil },
		InterfaceDrivers: func() (map[string]string, error) { return nil, errors.New("no ethtool") },
	}

	e := baseline()
	e.ExcludeDrivers = []string{"veth"}

	results := e.Validate(t.Context(), sys)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Result != FAIL || results[0].Property != "drivers" {
		t.Errorf("result = %+v, want a FAIL with property %q", results[0], "drivers")
	}
	if results[0].Err == nil || !strings.Contains(results[0].Err.Error(), "no ethtool") {
		t.Errorf("error = %v, want it to carry the lookup failure", results[0].Err)
	}
}

// Drivers are only looked up when a driver pattern actually needs them, so the
// common case does not pay a process per interface.
func TestEthtoolDefaultSkipsDriverLookupWithoutPatterns(t *testing.T) {
	lookedUp := false
	sys := &system.System{
		NewEthtool: func(_ context.Context, name string, _ *system.System, _ util.Config) system.Ethtool {
			iface := goodInterface()
			iface.name = name
			return &iface
		},
		InterfaceNames: func() ([]string, error) { return []string{"eth0"}, nil },
		InterfaceDrivers: func() (map[string]string, error) {
			lookedUp = true
			return nil, errors.New("drivers must not be needed here")
		},
	}

	results := baseline().Validate(t.Context(), sys)

	if lookedUp {
		t.Error("drivers were looked up although no exclude_drivers was set")
	}
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4", len(results))
	}
	for _, r := range results {
		if r.Result != SUCCESS {
			t.Errorf("result for %s.%s failed: %v", r.ResourceId, r.Property, r.Err)
		}
	}
}

// A baseline failure must name the offending interface in the resource id and
// the offending attribute in the property, which is the whole point of the
// default entry.
func TestEthtoolDefaultReportsOffendingInterfaceAndAttribute(t *testing.T) {
	slow := goodInterface()
	slow.speed = "1000Mb/s"
	sys := fakeEthtoolSystem([]string{"eth0", "eth1"}, map[string]fakeSysEthtool{
		"eth0": goodInterface(),
		"eth1": slow,
	})

	results := baseline().Validate(t.Context(), sys)

	var failed []TestResult
	for _, r := range results {
		if r.Result != SUCCESS {
			failed = append(failed, r)
		}
	}
	if len(failed) != 1 {
		t.Fatalf("got %d failures, want 1: %+v", len(failed), results)
	}
	if failed[0].ResourceId != "default: eth1" {
		t.Errorf("failing result resource id = %q, want %q", failed[0].ResourceId, "default: eth1")
	}
	if failed[0].Property != "speed" {
		t.Errorf("failing result property = %q, want %q", failed[0].Property, "speed")
	}
}

// An interface that has no ethtool attributes at all (a bridge, a veth pair)
// fails with a message naming the interface rather than passing silently.
func TestEthtoolDefaultFailsOnInterfaceWithoutReportedValues(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"virtio0"}, map[string]fakeSysEthtool{
		"virtio0": {},
	})

	results := baseline().Validate(t.Context(), sys)

	if len(results) != 4 {
		t.Fatalf("got %d results, want 4", len(results))
	}
	for _, r := range results {
		if r.Result != FAIL {
			t.Errorf("result for %s.%s = %d, want FAIL", r.ResourceId, r.Property, r.Result)
		}
		if r.Err == nil || !strings.Contains(r.Err.Error(), "virtio0") {
			t.Errorf("result for %s.%s has error %v, want one naming virtio0", r.ResourceId, r.Property, r.Err)
		}
	}
}

// A per-interface entry keeps the plain interface name as its id.
func TestEthtoolNamedEntryUsesInterfaceName(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0"}, map[string]fakeSysEthtool{"eth0": goodInterface()})

	e := &Ethtool{id: "eth0", Speed: "10000Mb/s", Duplex: "Full"}
	results := e.Validate(t.Context(), sys)

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if r.ResourceId != "eth0" {
			t.Errorf("result resource id = %q, want %q", r.ResourceId, "eth0")
		}
		if r.Result != SUCCESS {
			t.Errorf("result for eth0.%s failed: %v", r.Property, r.Err)
		}
	}
}

// skip on a default entry skips every expanded interface rather than only the
// entry itself.
func TestEthtoolDefaultSkipPropagates(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0", "eth1"}, map[string]fakeSysEthtool{
		"eth0": goodInterface(),
		"eth1": goodInterface(),
	})

	e := baseline()
	e.Skip = true
	results := e.Validate(t.Context(), sys)

	if len(results) != 8 {
		t.Fatalf("got %d results, want 8", len(results))
	}
	for _, r := range results {
		if !r.Skipped {
			t.Errorf("result for %s.%s was not skipped", r.ResourceId, r.Property)
		}
	}
}

// An entry with no attributes would expand to zero assertions, so it is
// rejected instead of passing while testing nothing.
func TestEthtoolWithoutAttributesFails(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0"}, map[string]fakeSysEthtool{"eth0": goodInterface()})

	for _, e := range []*Ethtool{{id: "eth0"}, {id: EthtoolDefaultID}} {
		results := e.Validate(t.Context(), sys)
		if len(results) != 1 {
			t.Fatalf("got %d results for %q, want 1", len(results), e.ID())
		}
		result := results[0]
		if result.Result != FAIL {
			t.Errorf("%q: result = %d, want FAIL", e.ID(), result.Result)
		}
		if result.Err == nil || !strings.Contains(result.Err.Error(), "no attributes set") {
			t.Errorf("%q: error = %v, want it to mention 'no attributes set'", e.ID(), result.Err)
		}
	}
}

// A default entry with exclude_interfaces but no attributes is still rejected,
// and is rejected once rather than once per interface.
func TestEthtoolDefaultWithoutAttributesFailsOnce(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0", "eth1", "eth2"}, map[string]fakeSysEthtool{})

	e := &Ethtool{id: EthtoolDefaultID, ExcludeInterfaces: []string{"lo"}}
	results := e.Validate(t.Context(), sys)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].ResourceId != EthtoolDefaultID {
		t.Errorf("result resource id = %q, want %q", results[0].ResourceId, EthtoolDefaultID)
	}
}

// A failure while enumerating interfaces is reported against the default entry.
func TestEthtoolDefaultInterfaceListError(t *testing.T) {
	sys := &system.System{
		NewEthtool: func(_ context.Context, name string, _ *system.System, _ util.Config) system.Ethtool {
			return &fakeSysEthtool{name: name}
		},
		InterfaceNames: func() ([]string, error) { return nil, errors.New("boom") },
	}

	results := baseline().Validate(t.Context(), sys)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Result != FAIL || results[0].Err == nil || !strings.Contains(results[0].Err.Error(), "boom") {
		t.Errorf("result = %+v, want a FAIL carrying the enumeration error", results[0])
	}
}

// A "default" key that also sets an explicit name is an ordinary entry for that
// interface, not the reserved baseline.
func TestEthtoolDefaultKeyWithNameIsNotBaseline(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0"}, map[string]fakeSysEthtool{"eth0": goodInterface()})

	e := &Ethtool{id: EthtoolDefaultID, Name: "eth0", Speed: "10000Mb/s"}
	if e.isDefault() {
		t.Fatal("an entry with an explicit name must not be treated as the baseline")
	}

	results := e.Validate(t.Context(), sys)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].ResourceId != "default: eth0" {
		t.Errorf("result resource id = %q, want %q", results[0].ResourceId, "default: eth0")
	}
}

// TestNewEthtoolBuildsAnEntryFromTheHost records the values reported by the
// system layer so that `goss add ethtool eth0` produces a file matching what
// the interface currently is.
func TestNewEthtoolBuildsAnEntryFromTheHost(t *testing.T) {
	e, err := NewEthtool(&fakeSysEthtool{
		name:           "eth0",
		speed:          "10000Mb/s",
		duplex:         "Full",
		rxChecksum:     "on",
		txChecksum:     "on",
		adaptiveRx:     "off",
		adaptiveTx:     "off",
		rxUsecs:        "3",
		txUsecs:        "3",
		rxRing:         "512",
		txRing:         "512",
		rxMiniRing:     "0",
		rxJumboRing:    "0",
		rxRingMax:      "4096",
		txRingMax:      "4096",
		rxMiniRingMax:  "0",
		rxJumboRingMax: "0",
		privFlags:      []string{"loopback: off", "rx-gro-hw: off"},
	}, util.Config{})
	if err != nil {
		t.Fatalf("NewEthtool returned error: %v", err)
	}

	if e.ID() != "eth0" {
		t.Errorf("ID() = %q, want %q", e.ID(), "eth0")
	}
	if e.Speed != "10000Mb/s" || e.Duplex != "Full" || e.RxChecksum != "on" || e.TxChecksum != "on" {
		t.Errorf("NewEthtool() = %+v, want the host's values", e)
	}
	if e.AdaptiveRx != "off" || e.AdaptiveTx != "off" || e.RxUsecs != "3" || e.TxUsecs != "3" {
		t.Errorf("NewEthtool() = %+v, want the host's coalescing values", e)
	}
	if e.RxRing != "512" || e.TxRing != "512" || e.RxRingMax != "4096" || e.TxRingMax != "4096" {
		t.Errorf("NewEthtool() = %+v, want the host's ring buffer sizes", e)
	}
	if e.RxMiniRing != "0" || e.RxMiniRingMax != "0" {
		t.Errorf("NewEthtool() = %+v, want the host's mini ring buffer sizes", e)
	}
	privFlags, ok := e.PrivFlags.([]string)
	if !ok || len(privFlags) != 2 || privFlags[0] != "loopback: off" {
		t.Errorf("NewEthtool() priv_flags = %#v, want the host's private flags", e.PrivFlags)
	}
}

// Values the host cannot report are left unset rather than recorded as a
// zero-value assertion that would always fail.
func TestNewEthtoolLeavesUnreportedAttributesUnset(t *testing.T) {
	e, err := NewEthtool(&fakeSysEthtool{name: "virtio0", speed: "10000Mb/s"}, util.Config{})
	if err != nil {
		t.Fatalf("NewEthtool returned error: %v", err)
	}

	if e.Speed != "10000Mb/s" {
		t.Errorf("Speed = %#v, want %q", e.Speed, "10000Mb/s")
	}
	if e.Duplex != nil || e.RxChecksum != nil || e.TxChecksum != nil {
		t.Errorf("unreported attributes should stay unset, got %+v", e)
	}
	if e.AdaptiveRx != nil || e.AdaptiveTx != nil || e.RxUsecs != nil || e.TxUsecs != nil {
		t.Errorf("unreported coalescing attributes should stay unset, got %+v", e)
	}
	if e.RxRing != nil || e.TxRing != nil || e.RxRingMax != nil || e.TxRingMax != nil {
		t.Errorf("unreported ring buffer attributes should stay unset, got %+v", e)
	}
	if e.RxMiniRing != nil || e.RxJumboRing != nil || e.RxMiniRingMax != nil || e.RxJumboRingMax != nil {
		t.Errorf("unreported mini and jumbo ring attributes should stay unset, got %+v", e)
	}
	if e.PrivFlags != nil {
		t.Errorf("unreported private flags should stay unset, got %+v", e.PrivFlags)
	}
}

// TestEthtoolResourceIsRegistered guards the CLI and gossfile paths that look
// the resource up by key.
func TestEthtoolResourceIsRegistered(t *testing.T) {
	res, ok := Resources()[EthtoolResourceKey]
	if !ok {
		t.Fatalf("resource %q is not registered", EthtoolResourceKey)
	}
	if res.TypeName() != EthtoolResourceName {
		t.Errorf("TypeName() = %q, want %q", res.TypeName(), EthtoolResourceName)
	}
}

// The coalescing attributes are expanded onto every interface just like the
// link settings, so one rule can pin adaptive coalescing and the usecs delays
// across the whole host.
func TestEthtoolDefaultExpandsCoalescingOntoEveryInterface(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0", "eth1"}, map[string]fakeSysEthtool{
		"eth0": goodInterface(),
		"eth1": goodInterface(),
	})

	results := coalescingBaseline().Validate(t.Context(), sys)

	if len(results) != 8 {
		t.Fatalf("got %d results, want 8 (4 attributes x 2 interfaces)", len(results))
	}
	for _, r := range results {
		if r.Result != SUCCESS {
			t.Errorf("result for %s.%s failed: %v", r.ResourceId, r.Property, r.Err)
		}
	}

	ids := resourceIDs(results)
	for _, want := range []string{"default: eth0", "default: eth1"} {
		if ids[want] != 4 {
			t.Errorf("resource id %q appeared %d times, want 4 (got ids %v)", want, ids[want], ids)
		}
	}

	properties := make(map[string]int)
	for _, r := range results {
		properties[r.Property]++
	}
	for _, want := range []string{"adaptive_rx", "adaptive_tx", "rx_usecs", "tx_usecs"} {
		if properties[want] != 2 {
			t.Errorf("property %q appeared %d times, want 2 (got properties %v)", want, properties[want], properties)
		}
	}
}

// A coalescing mismatch is reported against the interface and the attribute,
// not against the baseline entry as a whole.
func TestEthtoolDefaultReportsOffendingCoalescingAttribute(t *testing.T) {
	adaptive := goodInterface()
	adaptive.adaptiveRx = "on"
	sys := fakeEthtoolSystem([]string{"eth0", "eth1"}, map[string]fakeSysEthtool{
		"eth0": goodInterface(),
		"eth1": adaptive,
	})

	results := coalescingBaseline().Validate(t.Context(), sys)

	var failed []TestResult
	for _, r := range results {
		if r.Result != SUCCESS {
			failed = append(failed, r)
		}
	}
	if len(failed) != 1 {
		t.Fatalf("got %d failures, want 1: %+v", len(failed), results)
	}
	if failed[0].ResourceId != "default: eth1" {
		t.Errorf("failing result resource id = %q, want %q", failed[0].ResourceId, "default: eth1")
	}
	if failed[0].Property != "adaptive_rx" {
		t.Errorf("failing result property = %q, want %q", failed[0].Property, "adaptive_rx")
	}
}

// A coalescing value the host cannot report fails, naming the interface, rather
// than silently comparing against an empty string.
func TestEthtoolDefaultFailsOnUnreportedCoalescingValue(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0"}, map[string]fakeSysEthtool{
		"eth0": {speed: "10000Mb/s", duplex: "Full", rxChecksum: "on", txChecksum: "on"},
	})

	results := coalescingBaseline().Validate(t.Context(), sys)

	if len(results) != 4 {
		t.Fatalf("got %d results, want 4", len(results))
	}
	for _, r := range results {
		if r.Result != FAIL {
			t.Errorf("result for %s.%s = %d, want FAIL", r.ResourceId, r.Property, r.Result)
			continue
		}
		if r.Err == nil || !strings.Contains(r.Err.Error(), "eth0") {
			t.Errorf("result for %s.%s error = %v, want it to name the interface", r.ResourceId, r.Property, r.Err)
		}
	}
}

// The ring buffer attributes are expanded onto every interface, so one rule can
// pin the effective ring sizes and the driver's ceiling across the whole host.
func TestEthtoolDefaultExpandsRingOntoEveryInterface(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0", "eth1"}, map[string]fakeSysEthtool{
		"eth0": goodInterface(),
		"eth1": goodInterface(),
	})

	results := ringBaseline().Validate(t.Context(), sys)

	if len(results) != 16 {
		t.Fatalf("got %d results, want 16 (8 attributes x 2 interfaces)", len(results))
	}
	for _, r := range results {
		if r.Result != SUCCESS {
			t.Errorf("result for %s.%s failed: %v", r.ResourceId, r.Property, r.Err)
		}
	}

	ids := resourceIDs(results)
	for _, want := range []string{"default: eth0", "default: eth1"} {
		if ids[want] != 8 {
			t.Errorf("resource id %q appeared %d times, want 8 (got ids %v)", want, ids[want], ids)
		}
	}

	properties := make(map[string]int)
	for _, r := range results {
		properties[r.Property]++
	}
	for _, want := range []string{"rx_ring", "tx_ring", "rx_mini_ring", "rx_jumbo_ring", "rx_ring_max", "tx_ring_max", "rx_mini_ring_max", "rx_jumbo_ring_max"} {
		if properties[want] != 2 {
			t.Errorf("property %q appeared %d times, want 2 (got properties %v)", want, properties[want], properties)
		}
	}
}

// A ring buffer mismatch is reported against the interface and the attribute,
// not against the baseline entry as a whole.
func TestEthtoolDefaultReportsOffendingRingAttribute(t *testing.T) {
	shallow := goodInterface()
	shallow.rxRing = "256"
	sys := fakeEthtoolSystem([]string{"eth0", "eth1"}, map[string]fakeSysEthtool{
		"eth0": goodInterface(),
		"eth1": shallow,
	})

	results := ringBaseline().Validate(t.Context(), sys)

	var failed []TestResult
	for _, r := range results {
		if r.Result != SUCCESS {
			failed = append(failed, r)
		}
	}
	if len(failed) != 1 {
		t.Fatalf("got %d failures, want 1: %+v", len(failed), results)
	}
	if failed[0].ResourceId != "default: eth1" {
		t.Errorf("failing result resource id = %q, want %q", failed[0].ResourceId, "default: eth1")
	}
	if failed[0].Property != "rx_ring" {
		t.Errorf("failing result property = %q, want %q", failed[0].Property, "rx_ring")
	}
}

// A ring buffer size the host cannot report fails, naming the interface, rather
// than silently comparing against an empty string.
func TestEthtoolDefaultFailsOnUnreportedRingValue(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0"}, map[string]fakeSysEthtool{
		"eth0": {speed: "10000Mb/s", duplex: "Full", rxChecksum: "on", txChecksum: "on"},
	})

	results := ringBaseline().Validate(t.Context(), sys)

	if len(results) != 8 {
		t.Fatalf("got %d results, want 8", len(results))
	}
	for _, r := range results {
		if r.Result != FAIL {
			t.Errorf("result for %s.%s = %d, want FAIL", r.ResourceId, r.Property, r.Result)
			continue
		}
		if r.Err == nil || !strings.Contains(r.Err.Error(), "eth0") {
			t.Errorf("result for %s.%s error = %v, want it to name the interface", r.ResourceId, r.Property, r.Err)
		}
	}
}

// An expected private flag list has to be contained in what the interface
// reports, so asserting two of three flags passes and a flag the interface does
// not have fails.
func TestEthtoolPrivFlagsAreMatchedAsASubset(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0"}, map[string]fakeSysEthtool{"eth0": goodInterface()})

	results := privFlagsBaseline().Validate(t.Context(), sys)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1: %+v", len(results), results)
	}
	for _, r := range results {
		if r.Result != SUCCESS {
			t.Errorf("result for %s.%s failed: %v", r.ResourceId, r.Property, r.Err)
		}
	}

	// A flag the interface does not report at all is a failure.
	missing := privFlagsBaseline()
	missing.PrivFlags = []any{"tls-hw-tx-offload: on"}
	results = missing.Validate(t.Context(), sys)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Result != FAIL {
		t.Errorf("result for a missing private flag = %d, want FAIL", results[0].Result)
	}
	if results[0].ResourceId != "default: eth0" {
		t.Errorf("failing result resource id = %q, want %q", results[0].ResourceId, "default: eth0")
	}
	if results[0].Property != "priv_flags" {
		t.Errorf("failing result property = %q, want %q", results[0].Property, "priv_flags")
	}
}

// An interface that reports no private flags at all fails, naming the interface.
func TestEthtoolDefaultFailsOnUnreportedPrivFlags(t *testing.T) {
	sys := fakeEthtoolSystem([]string{"eth0"}, map[string]fakeSysEthtool{
		"eth0": {speed: "10000Mb/s", duplex: "Full", rxChecksum: "on", txChecksum: "on"},
	})

	results := privFlagsBaseline().Validate(t.Context(), sys)

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Result != FAIL {
		t.Errorf("result = %d, want FAIL", results[0].Result)
	}
	if results[0].Err == nil || !strings.Contains(results[0].Err.Error(), "eth0") {
		t.Errorf("error = %v, want it to name the interface", results[0].Err)
	}
}
