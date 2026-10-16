package system

import (
	"context"
	"errors"
	"net"
	"sort"

	"github.com/goss-org/goss/util"
)

// Ethtool exposes the link settings, offload features, interrupt coalescing
// settings, ring buffers and private flags that ethtool(8) reports for a single
// network interface.
//
// Every getter returns the value exactly as ethtool printed it, so a gossfile
// can match on it directly (e.g. "10000Mb/s", "Full", "on", "512").
type Ethtool interface {
	Name() string
	Exists() (bool, error)
	Speed() (string, error)
	Duplex() (string, error)
	RxChecksum() (string, error)
	TxChecksum() (string, error)
	// AdaptiveRx, AdaptiveTx, RxUsecs and TxUsecs are the interrupt coalescing
	// settings reported by "ethtool -c".
	AdaptiveRx() (string, error)
	AdaptiveTx() (string, error)
	RxUsecs() (string, error)
	TxUsecs() (string, error)
	// RingRx, RingTx, RingRxMini and RingRxJumbo are the ring buffer sizes the
	// interface currently uses, and RingRxMax, RingTxMax, RingRxMiniMax and
	// RingRxJumboMax the pre-set maximums, all reported by "ethtool -g".
	RingRx() (string, error)
	RingTx() (string, error)
	RingRxMini() (string, error)
	RingRxJumbo() (string, error)
	RingRxMax() (string, error)
	RingTxMax() (string, error)
	RingRxMiniMax() (string, error)
	RingRxJumboMax() (string, error)
	// PrivFlags are the private flags the driver exposes for the interface
	// ("ethtool --show-priv-flags"), as sorted "name: value" pairs.
	PrivFlags() ([]string, error)
	// Driver is the kernel driver ethtool attributes to the interface
	// ("ethtool -i"), such as "ixgbe", "virtio_net" or "veth".
	Driver() (string, error)
}

var (
	// ErrEthtoolUnsupported is returned on platforms that have no ethtool(8).
	ErrEthtoolUnsupported = errors.New("ethtool resource is only supported on Linux")
	// ErrEthtoolNotFound is returned when the ethtool binary is not on the PATH.
	ErrEthtoolNotFound = errors.New("ethtool command not found in PATH")
)

// InterfaceNames returns the names of every network interface on the host,
// sorted lexicographically.
func InterfaceNames() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		names = append(names, iface.Name)
	}
	sort.Strings(names)
	return names, nil
}

// InterfaceDrivers maps every interface on the host to the kernel driver for it,
// so that a gossfile can exempt whole classes of interface ("veth", "bridge")
// instead of naming them one by one.
//
// Interfaces whose driver cannot be determined -- "lo", or anything with no
// driver information -- are left out of the map rather than reported with an
// empty name, so that a driver pattern can never match them by accident.
func InterfaceDrivers() (map[string]string, error) {
	names, err := InterfaceNames()
	if err != nil {
		return nil, err
	}

	drivers := make(map[string]string, len(names))
	for _, name := range names {
		// An interface backed by a device carries its driver in sysfs, which
		// costs no process; ethtool is only asked about the rest.
		if driver, ok := sysfsDriver(name); ok {
			drivers[name] = driver
			continue
		}

		driver, err := NewDefEthtool(context.Background(), name, nil, util.Config{}).Driver()
		if err != nil {
			// A host without ethtool cannot answer this question at all. An
			// individual interface with no driver information just does not
			// participate in driver exclusion.
			if errors.Is(err, ErrEthtoolNotFound) || errors.Is(err, ErrEthtoolUnsupported) {
				return nil, err
			}
			continue
		}
		drivers[name] = driver
	}
	return drivers, nil
}
