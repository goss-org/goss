//go:build linux

package system

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goss-org/goss/util"
)

// defEthtool reads link settings, offload features, the driver, the coalescing
// settings, the ring buffers and the private flags by shelling out to
// ethtool(8). Every read is cached, so a resource only pays for what it
// actually asserts on: two process spawns for the settings and features, one
// more for the driver and one more for each of the coalescing settings, the
// ring buffers and the private flags.
type defEthtool struct {
	name     string
	loaded   bool
	exists   bool
	err      error
	settings map[string]string
	features map[string]string

	driverLoaded bool
	driver       string
	driverErr    error

	coalesceLoaded bool
	coalesce       map[string]string
	coalesceErr    error

	ringLoaded bool
	ring       map[string]string
	ringErr    error

	privFlagsLoaded bool
	privFlags       []string
	privFlagsErr    error
}

func NewDefEthtool(_ context.Context, name string, _ *System, _ util.Config) Ethtool {
	return &defEthtool{name: name}
}

func (e *defEthtool) setup() error {
	if e.loaded {
		return e.err
	}
	e.loaded = true

	if _, err := exec.LookPath("ethtool"); err != nil {
		e.err = ErrEthtoolNotFound
		return e.err
	}

	// A missing interface is a failed assertion, not a runtime error, so it is
	// reported through Exists() rather than through the error channel.
	if _, err := net.InterfaceByName(e.name); err != nil {
		return nil
	}
	e.exists = true

	out, err := runEthtool(e.name)
	if err != nil {
		e.err = err
		return e.err
	}
	e.settings = parseEthtoolFields(out)

	out, err = runEthtool("-k", e.name)
	if err != nil {
		e.err = err
		return e.err
	}
	e.features = parseEthtoolFields(out)

	return nil
}

func runEthtool(args ...string) (string, error) {
	cmd := exec.Command("ethtool", args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("ethtool %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}

// sysfsNetPath is where the kernel exposes the network interfaces. It is a
// variable so that tests can point it at a fixture.
var sysfsNetPath = "/sys/class/net"

// sysfsDriver reads the driver name out of the sysfs driver symlink, which is
// the same information "ethtool -i" reports but without spawning a process.
// Interfaces that are not backed by a device -- "lo", "veth", "br0" -- have no
// such symlink and report false, so the caller falls back to ethtool.
func sysfsDriver(name string) (string, bool) {
	target, err := os.Readlink(filepath.Join(sysfsNetPath, name, "device", "driver"))
	if err != nil {
		return "", false
	}
	driver := path.Base(target)
	if driver == "" || driver == "." || driver == "/" {
		return "", false
	}
	return driver, true
}

// parseEthtoolFields turns ethtool's "Key: value" output into a map. It is
// deliberately tolerant: any line without a separator is skipped, and only the
// first separator splits, so values that contain a colon survive intact.
func parseEthtoolFields(out string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		fields[key] = strings.TrimSpace(value)
	}
	return fields
}

// parseEthtoolCoalesce reads "ethtool -c" output. It mostly fits the
// "Key: value" shape, except adaptive coalescing, which ethtool reports as a
// single "Adaptive RX: off  TX: off" line; the two values are split out into
// adaptive_rx and adaptive_tx so that each can be asserted on its own.
func parseEthtoolCoalesce(out string) map[string]string {
	fields := parseEthtoolFields(out)

	adaptive, ok := fields["Adaptive RX"]
	if !ok {
		return fields
	}
	rx, tx, found := strings.Cut(adaptive, "TX:")
	fields["adaptive_rx"] = strings.TrimSpace(rx)
	if found {
		fields["adaptive_tx"] = strings.TrimSpace(tx)
	}
	return fields
}

// parseEthtoolRing reads "ethtool -g" output, which prints the same four
// labels twice: once under the pre-set maximums and once under the current
// hardware settings. The maximums are kept under a _max suffix, so a rule can
// assert the effective size, the ceiling the driver allows, or both.
func parseEthtoolRing(out string) map[string]string {
	fields := make(map[string]string)
	maxSection := false

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		switch line {
		case "Pre-set maximums:":
			maxSection = true
			continue
		case "Current hardware settings:":
			maxSection = false
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = ringField(strings.TrimSpace(key))
		if key == "" {
			continue
		}
		if maxSection {
			key += "_max"
		}
		fields[key] = strings.TrimSpace(value)
	}
	return fields
}

// ringField maps the labels ethtool prints for a ring buffer onto attribute
// names. Labels we do not know are dropped rather than invented.
func ringField(label string) string {
	switch label {
	case "RX":
		return "rx_ring"
	case "TX":
		return "tx_ring"
	case "RX Mini":
		return "rx_mini_ring"
	case "RX Jumbo":
		return "rx_jumbo_ring"
	}
	return ""
}

// parsePrivFlags reads "ethtool --show-priv-flags" output as a sorted list of
// "name: value" pairs. A list keeps goss's default matcher, in which the
// expected list has to be contained in the actual one, so a private flag that a
// newer kernel adds does not break a rule that never mentioned it.
func parsePrivFlags(out string) []string {
	fields := parseEthtoolFields(out)

	flags := make([]string, 0, len(fields))
	for name, value := range fields {
		if value == "" {
			// "Private flags for eth0:" is a header, not a flag.
			continue
		}
		flags = append(flags, fmt.Sprintf("%s: %s", name, value))
	}
	sort.Strings(flags)
	return flags
}

func (e *defEthtool) Name() string { return e.name }

func (e *defEthtool) Exists() (bool, error) {
	if err := e.setup(); err != nil {
		return false, err
	}
	return e.exists, nil
}

func (e *defEthtool) Speed() (string, error) {
	return e.field(e.settings, "Speed", e.setup())
}

func (e *defEthtool) Duplex() (string, error) {
	return e.field(e.settings, "Duplex", e.setup())
}

func (e *defEthtool) RxChecksum() (string, error) {
	return e.field(e.features, "rx-checksumming", e.setup())
}

func (e *defEthtool) TxChecksum() (string, error) {
	return e.field(e.features, "tx-checksumming", e.setup())
}

// AdaptiveRx and AdaptiveTx report whether the interface adapts its interrupt
// coalescing to the traffic. ethtool only labels the first of the pair, so the
// TX value is taken from the same line.
func (e *defEthtool) AdaptiveRx() (string, error) {
	fields, err := e.coalesceSettings()
	return e.field(fields, "adaptive_rx", err)
}

func (e *defEthtool) AdaptiveTx() (string, error) {
	fields, err := e.coalesceSettings()
	return e.field(fields, "adaptive_tx", err)
}

// RxUsecs and TxUsecs are the interrupt coalescing delays, in microseconds.
func (e *defEthtool) RxUsecs() (string, error) {
	fields, err := e.coalesceSettings()
	return e.field(fields, "rx-usecs", err)
}

func (e *defEthtool) TxUsecs() (string, error) {
	fields, err := e.coalesceSettings()
	return e.field(fields, "tx-usecs", err)
}

// RingRx, RingTx, RingRxMini and RingRxJumbo are the ring buffer sizes the
// interface is currently using ("ethtool -g"). RingRxMax, RingTxMax,
// RingRxMiniMax and RingRxJumboMax are the pre-set maximums the driver allows,
// which is what a rule on a host-wide baseline usually wants to pin down.
func (e *defEthtool) RingRx() (string, error) {
	fields, err := e.ringSettings()
	return e.field(fields, "rx_ring", err)
}

func (e *defEthtool) RingTx() (string, error) {
	fields, err := e.ringSettings()
	return e.field(fields, "tx_ring", err)
}

func (e *defEthtool) RingRxMini() (string, error) {
	fields, err := e.ringSettings()
	return e.field(fields, "rx_mini_ring", err)
}

func (e *defEthtool) RingRxJumbo() (string, error) {
	fields, err := e.ringSettings()
	return e.field(fields, "rx_jumbo_ring", err)
}

func (e *defEthtool) RingRxMax() (string, error) {
	fields, err := e.ringSettings()
	return e.field(fields, "rx_ring_max", err)
}

func (e *defEthtool) RingTxMax() (string, error) {
	fields, err := e.ringSettings()
	return e.field(fields, "tx_ring_max", err)
}

func (e *defEthtool) RingRxMiniMax() (string, error) {
	fields, err := e.ringSettings()
	return e.field(fields, "rx_mini_ring_max", err)
}

func (e *defEthtool) RingRxJumboMax() (string, error) {
	fields, err := e.ringSettings()
	return e.field(fields, "rx_jumbo_ring_max", err)
}

// PrivFlags are the private flags the driver exposes for the interface
// ("ethtool --show-priv-flags"), as sorted "name: value" pairs.
func (e *defEthtool) PrivFlags() ([]string, error) {
	if err := e.setup(); err != nil {
		return nil, err
	}
	if e.privFlagsLoaded {
		return e.privFlags, e.privFlagsErr
	}
	e.privFlagsLoaded = true

	if !e.exists {
		e.privFlagsErr = fmt.Errorf("interface %s does not exist", e.name)
		return nil, e.privFlagsErr
	}

	out, err := runEthtool("--show-priv-flags", e.name)
	if err != nil {
		e.privFlagsErr = err
		return nil, e.privFlagsErr
	}
	e.privFlags = parsePrivFlags(out)
	return e.privFlags, nil
}

// Driver reads the kernel driver from "ethtool -i", which reports it for
// physical NICs and for most virtual ones ("veth", "bridge", "tun"). Interfaces
// ethtool has no driver information for -- "lo" for instance -- return an error
// naming the interface.
func (e *defEthtool) Driver() (string, error) {
	if err := e.setup(); err != nil {
		return "", err
	}
	if e.driverLoaded {
		return e.driver, e.driverErr
	}
	e.driverLoaded = true

	if !e.exists {
		e.driverErr = fmt.Errorf("interface %s does not exist", e.name)
		return "", e.driverErr
	}

	out, err := runEthtool("-i", e.name)
	if err != nil {
		e.driverErr = err
		return "", e.driverErr
	}

	driver, ok := parseEthtoolFields(out)["driver"]
	if !ok {
		e.driverErr = fmt.Errorf("ethtool did not report %q for interface %s", "driver", e.name)
		return "", e.driverErr
	}
	e.driver = driver
	return e.driver, nil
}

// coalesceSettings reads "ethtool -c" once, and only when a coalescing
// attribute is actually asserted on.
func (e *defEthtool) coalesceSettings() (map[string]string, error) {
	if err := e.setup(); err != nil {
		return nil, err
	}
	if e.coalesceLoaded {
		return e.coalesce, e.coalesceErr
	}
	e.coalesceLoaded = true

	if !e.exists {
		e.coalesceErr = fmt.Errorf("interface %s does not exist", e.name)
		return nil, e.coalesceErr
	}

	out, err := runEthtool("-c", e.name)
	if err != nil {
		e.coalesceErr = err
		return nil, e.coalesceErr
	}
	e.coalesce = parseEthtoolCoalesce(out)
	return e.coalesce, nil
}

// ringSettings reads "ethtool -g" once, and only when a ring buffer attribute
// is actually asserted on.
func (e *defEthtool) ringSettings() (map[string]string, error) {
	if err := e.setup(); err != nil {
		return nil, err
	}
	if e.ringLoaded {
		return e.ring, e.ringErr
	}
	e.ringLoaded = true

	if !e.exists {
		e.ringErr = fmt.Errorf("interface %s does not exist", e.name)
		return nil, e.ringErr
	}

	out, err := runEthtool("-g", e.name)
	if err != nil {
		e.ringErr = err
		return nil, e.ringErr
	}
	e.ring = parseEthtoolRing(out)
	return e.ring, nil
}

// field reads one value, turning the two expected failure modes into a message
// that names the interface, so a baseline rule over many interfaces stays
// diagnosable.
func (e *defEthtool) field(fields map[string]string, key string, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if !e.exists {
		return "", fmt.Errorf("interface %s does not exist", e.name)
	}
	value, ok := fields[key]
	if !ok {
		return "", fmt.Errorf("ethtool did not report %q for interface %s", key, e.name)
	}
	return value, nil
}
