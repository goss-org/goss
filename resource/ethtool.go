package resource

import (
	"context"
	"fmt"
	"time"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

// EthtoolDefaultID is the reserved resource key whose attributes are expanded
// onto every network interface on the host, minus exclude_interfaces.
const EthtoolDefaultID = "default"

// Ethtool validates the link settings, offload features, interrupt coalescing
// settings, ring buffers and private flags that ethtool(8) reports for a
// network interface.
//
// Besides the usual per-interface entry, a single "default" entry carries a
// baseline that every interface must satisfy:
//
//	ethtool:
//	  default:
//	    speed: "10000Mb/s"
//	    duplex: "Full"
//	    rx_checksum: "on"
//	    tx_checksum: "on"
//	    exclude_interfaces:
//	      - lo
//	      - docker0
//	      - veth*
//	    exclude_drivers:
//	      - veth
//	      - bridge
//	    adaptive_rx: "off"
//	    adaptive_tx: "off"
//	    rx_usecs: 0
//	    tx_usecs: 0
//	    rx_ring: 512
//	    tx_ring: 512
//	    rx_ring_max: 4096
//	    tx_ring_max: 4096
//	    priv_flags:
//	      - "rx-gro-hw: off"
//
// A default entry is expanded while validating, so a failure names both the
// offending interface and the attribute that did not match.
type Ethtool struct {
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Meta              meta     `json:"meta,omitempty" yaml:"meta,omitempty"`
	id                string   `json:"-" yaml:"-"`
	Name              string   `json:"name,omitempty" yaml:"name,omitempty"`
	Speed             matcher  `json:"speed,omitempty" yaml:"speed,omitempty"`
	Duplex            matcher  `json:"duplex,omitempty" yaml:"duplex,omitempty"`
	RxChecksum        matcher  `json:"rx_checksum,omitempty" yaml:"rx_checksum,omitempty"`
	TxChecksum        matcher  `json:"tx_checksum,omitempty" yaml:"tx_checksum,omitempty"`
	AdaptiveRx        matcher  `json:"adaptive_rx,omitempty" yaml:"adaptive_rx,omitempty"`
	AdaptiveTx        matcher  `json:"adaptive_tx,omitempty" yaml:"adaptive_tx,omitempty"`
	RxUsecs           matcher  `json:"rx_usecs,omitempty" yaml:"rx_usecs,omitempty"`
	TxUsecs           matcher  `json:"tx_usecs,omitempty" yaml:"tx_usecs,omitempty"`
	RxRing            matcher  `json:"rx_ring,omitempty" yaml:"rx_ring,omitempty"`
	TxRing            matcher  `json:"tx_ring,omitempty" yaml:"tx_ring,omitempty"`
	RxMiniRing        matcher  `json:"rx_mini_ring,omitempty" yaml:"rx_mini_ring,omitempty"`
	RxJumboRing       matcher  `json:"rx_jumbo_ring,omitempty" yaml:"rx_jumbo_ring,omitempty"`
	RxRingMax         matcher  `json:"rx_ring_max,omitempty" yaml:"rx_ring_max,omitempty"`
	TxRingMax         matcher  `json:"tx_ring_max,omitempty" yaml:"tx_ring_max,omitempty"`
	RxMiniRingMax     matcher  `json:"rx_mini_ring_max,omitempty" yaml:"rx_mini_ring_max,omitempty"`
	RxJumboRingMax    matcher  `json:"rx_jumbo_ring_max,omitempty" yaml:"rx_jumbo_ring_max,omitempty"`
	PrivFlags         matcher  `json:"priv_flags,omitempty" yaml:"priv_flags,omitempty"`
	ExcludeInterfaces []string `json:"exclude_interfaces,omitempty" yaml:"exclude_interfaces,omitempty"`
	ExcludeDrivers    []string `json:"exclude_drivers,omitempty" yaml:"exclude_drivers,omitempty"`
	Skip              bool     `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	EthtoolResourceKey  = "ethtool"
	EthtoolResourceName = "Ethtool"
)

func init() {
	registerResource(EthtoolResourceKey, &Ethtool{})
}

func (e *Ethtool) ID() string {
	if e.Name != "" && e.Name != e.id {
		return fmt.Sprintf("%s: %s", e.id, e.Name)
	}
	return e.id
}
func (e *Ethtool) SetID(id string)  { e.id = id }
func (e *Ethtool) SetSkip()         { e.Skip = true }
func (e *Ethtool) TypeKey() string  { return EthtoolResourceKey }
func (e *Ethtool) TypeName() string { return EthtoolResourceName }

func (e *Ethtool) GetTitle() string { return e.Title }
func (e *Ethtool) GetMeta() meta    { return e.Meta }
func (e *Ethtool) GetName() string {
	if e.Name != "" {
		return e.Name
	}
	return e.id
}

// isDefault reports whether this entry is the reserved baseline entry. A
// "default" key that also sets an explicit name is an ordinary entry for that
// interface, so a host with an interface literally called "default" still works.
func (e *Ethtool) isDefault() bool {
	return e.id == EthtoolDefaultID && e.Name == ""
}

func (e *Ethtool) Validate(ctx context.Context, sys *system.System) []TestResult {
	if !e.isDefault() {
		return e.validateInterface(ctx, sys, e.GetName(), e.Skip)
	}
	if err := e.attributesMissing(); err != nil {
		return []TestResult{ethtoolErrorResult(e, "attributes", err)}
	}
	ctx = context.WithValue(ctx, idKey{}, e.ID())
	return e.validateAllInterfaces(ctx, sys)
}

// validateAllInterfaces expands the baseline onto every interface reported by
// the host, skipping the ones named by exclude_interfaces and the ones driven by
// a driver named by exclude_drivers. Patterns are matched with the same glob
// rules as gossfile keys, so "veth*" skips every veth pair, and a driver pattern
// of "veth" skips every interface the kernel says is driven by veth.
func (e *Ethtool) validateAllInterfaces(ctx context.Context, sys *system.System) []TestResult {
	names, err := interfaceNames(sys)
	if err != nil {
		return []TestResult{ethtoolErrorResult(e, "interfaces", err)}
	}

	// The driver of each interface is only looked up when a driver pattern needs
	// it, because that costs a process per interface.
	var drivers map[string]string
	if len(e.ExcludeDrivers) > 0 {
		drivers, err = interfaceDrivers(sys)
		if err != nil {
			return []TestResult{ethtoolErrorResult(e, "drivers", err)}
		}
	}

	var results []TestResult
	for _, name := range names {
		if contains(e.ExcludeInterfaces, name) {
			continue
		}
		// An interface whose driver is unknown is never excluded by a driver
		// pattern; name it in exclude_interfaces instead.
		if driver, ok := drivers[name]; ok && contains(e.ExcludeDrivers, driver) {
			continue
		}
		results = append(results, e.validateInterface(ctx, sys, name, e.Skip)...)
	}
	return results
}

// validateInterface evaluates the configured attributes against one interface.
// name is the interface being read; it differs from the resource id only when a
// baseline entry has been expanded, and the id then becomes "default: <name>"
// so that failures point at the offending interface.
func (e *Ethtool) validateInterface(ctx context.Context, sys *system.System, name string, skip bool) []TestResult {
	view := *e
	view.Name = name
	ctx = context.WithValue(ctx, idKey{}, view.ID())

	if err := view.attributesMissing(); err != nil {
		return []TestResult{ethtoolErrorResult(&view, "attributes", err)}
	}

	sysEthtool := sys.NewEthtool(ctx, name, sys, util.Config{})

	var results []TestResult
	if view.Speed != nil {
		results = append(results, ValidateValue(&view, "speed", view.Speed, sysEthtool.Speed, skip))
	}
	if view.Duplex != nil {
		results = append(results, ValidateValue(&view, "duplex", view.Duplex, sysEthtool.Duplex, skip))
	}
	if view.RxChecksum != nil {
		results = append(results, ValidateValue(&view, "rx_checksum", view.RxChecksum, sysEthtool.RxChecksum, skip))
	}
	if view.TxChecksum != nil {
		results = append(results, ValidateValue(&view, "tx_checksum", view.TxChecksum, sysEthtool.TxChecksum, skip))
	}
	if view.AdaptiveRx != nil {
		results = append(results, ValidateValue(&view, "adaptive_rx", view.AdaptiveRx, sysEthtool.AdaptiveRx, skip))
	}
	if view.AdaptiveTx != nil {
		results = append(results, ValidateValue(&view, "adaptive_tx", view.AdaptiveTx, sysEthtool.AdaptiveTx, skip))
	}
	if view.RxUsecs != nil {
		results = append(results, ValidateValue(&view, "rx_usecs", view.RxUsecs, sysEthtool.RxUsecs, skip))
	}
	if view.TxUsecs != nil {
		results = append(results, ValidateValue(&view, "tx_usecs", view.TxUsecs, sysEthtool.TxUsecs, skip))
	}
	if view.RxRing != nil {
		results = append(results, ValidateValue(&view, "rx_ring", view.RxRing, sysEthtool.RingRx, skip))
	}
	if view.TxRing != nil {
		results = append(results, ValidateValue(&view, "tx_ring", view.TxRing, sysEthtool.RingTx, skip))
	}
	if view.RxMiniRing != nil {
		results = append(results, ValidateValue(&view, "rx_mini_ring", view.RxMiniRing, sysEthtool.RingRxMini, skip))
	}
	if view.RxJumboRing != nil {
		results = append(results, ValidateValue(&view, "rx_jumbo_ring", view.RxJumboRing, sysEthtool.RingRxJumbo, skip))
	}
	if view.RxRingMax != nil {
		results = append(results, ValidateValue(&view, "rx_ring_max", view.RxRingMax, sysEthtool.RingRxMax, skip))
	}
	if view.TxRingMax != nil {
		results = append(results, ValidateValue(&view, "tx_ring_max", view.TxRingMax, sysEthtool.RingTxMax, skip))
	}
	if view.RxMiniRingMax != nil {
		results = append(results, ValidateValue(&view, "rx_mini_ring_max", view.RxMiniRingMax, sysEthtool.RingRxMiniMax, skip))
	}
	if view.RxJumboRingMax != nil {
		results = append(results, ValidateValue(&view, "rx_jumbo_ring_max", view.RxJumboRingMax, sysEthtool.RingRxJumboMax, skip))
	}
	if view.PrivFlags != nil {
		results = append(results, ValidateValue(&view, "priv_flags", view.PrivFlags, sysEthtool.PrivFlags, skip))
	}
	return results
}

// attributesMissing rejects an entry that carries no matcher at all. Such an
// entry would expand to zero assertions and pass while testing nothing.
func (e *Ethtool) attributesMissing() error {
	if e.Speed == nil && e.Duplex == nil && e.RxChecksum == nil && e.TxChecksum == nil &&
		e.AdaptiveRx == nil && e.AdaptiveTx == nil && e.RxUsecs == nil && e.TxUsecs == nil &&
		e.RxRing == nil && e.TxRing == nil && e.RxMiniRing == nil && e.RxJumboRing == nil &&
		e.RxRingMax == nil && e.TxRingMax == nil && e.RxMiniRingMax == nil && e.RxJumboRingMax == nil &&
		e.PrivFlags == nil {
		return fmt.Errorf("no attributes set, expected at least one of speed, duplex, rx_checksum, tx_checksum, " +
			"adaptive_rx, adaptive_tx, rx_usecs, tx_usecs, rx_ring, tx_ring, rx_mini_ring, rx_jumbo_ring, " +
			"rx_ring_max, tx_ring_max, rx_mini_ring_max, rx_jumbo_ring_max, priv_flags")
	}
	return nil
}

func ethtoolErrorResult(res ResourceRead, property string, err error) TestResult {
	startTime := time.Now()
	endTime := time.Now()
	return TestResult{
		Result:       FAIL,
		ResourceType: EthtoolResourceName,
		ResourceId:   res.ID(),
		Title:        res.GetTitle(),
		Meta:         res.GetMeta(),
		Property:     property,
		Err:          toValidateError(err),
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
	}
}

// interfaceNames prefers the System hook so that embedders and tests can supply
// their own interface list.
func interfaceNames(sys *system.System) ([]string, error) {
	if sys.InterfaceNames != nil {
		return sys.InterfaceNames()
	}
	return system.InterfaceNames()
}

// interfaceDrivers prefers the System hook so that embedders and tests can
// supply their own interface to driver mapping.
func interfaceDrivers(sys *system.System) (map[string]string, error) {
	if sys.InterfaceDrivers != nil {
		return sys.InterfaceDrivers()
	}
	return system.InterfaceDrivers()
}

func NewEthtool(sysEthtool system.Ethtool, config util.Config) (*Ethtool, error) {
	name := sysEthtool.Name()
	e := &Ethtool{id: name}
	if !contains(config.IgnoreList, "speed") {
		if speed, err := sysEthtool.Speed(); err == nil {
			e.Speed = speed
		}
	}
	if !contains(config.IgnoreList, "duplex") {
		if duplex, err := sysEthtool.Duplex(); err == nil {
			e.Duplex = duplex
		}
	}
	if !contains(config.IgnoreList, "rx_checksum") {
		if rxChecksum, err := sysEthtool.RxChecksum(); err == nil {
			e.RxChecksum = rxChecksum
		}
	}
	if !contains(config.IgnoreList, "tx_checksum") {
		if txChecksum, err := sysEthtool.TxChecksum(); err == nil {
			e.TxChecksum = txChecksum
		}
	}
	if !contains(config.IgnoreList, "adaptive_rx") {
		if adaptiveRx, err := sysEthtool.AdaptiveRx(); err == nil {
			e.AdaptiveRx = adaptiveRx
		}
	}
	if !contains(config.IgnoreList, "adaptive_tx") {
		if adaptiveTx, err := sysEthtool.AdaptiveTx(); err == nil {
			e.AdaptiveTx = adaptiveTx
		}
	}
	if !contains(config.IgnoreList, "rx_usecs") {
		if rxUsecs, err := sysEthtool.RxUsecs(); err == nil {
			e.RxUsecs = rxUsecs
		}
	}
	if !contains(config.IgnoreList, "tx_usecs") {
		if txUsecs, err := sysEthtool.TxUsecs(); err == nil {
			e.TxUsecs = txUsecs
		}
	}
	if !contains(config.IgnoreList, "rx_ring") {
		if rxRing, err := sysEthtool.RingRx(); err == nil {
			e.RxRing = rxRing
		}
	}
	if !contains(config.IgnoreList, "tx_ring") {
		if txRing, err := sysEthtool.RingTx(); err == nil {
			e.TxRing = txRing
		}
	}
	if !contains(config.IgnoreList, "rx_mini_ring") {
		if rxMiniRing, err := sysEthtool.RingRxMini(); err == nil {
			e.RxMiniRing = rxMiniRing
		}
	}
	if !contains(config.IgnoreList, "rx_jumbo_ring") {
		if rxJumboRing, err := sysEthtool.RingRxJumbo(); err == nil {
			e.RxJumboRing = rxJumboRing
		}
	}
	if !contains(config.IgnoreList, "rx_ring_max") {
		if rxRingMax, err := sysEthtool.RingRxMax(); err == nil {
			e.RxRingMax = rxRingMax
		}
	}
	if !contains(config.IgnoreList, "tx_ring_max") {
		if txRingMax, err := sysEthtool.RingTxMax(); err == nil {
			e.TxRingMax = txRingMax
		}
	}
	if !contains(config.IgnoreList, "rx_mini_ring_max") {
		if rxMiniRingMax, err := sysEthtool.RingRxMiniMax(); err == nil {
			e.RxMiniRingMax = rxMiniRingMax
		}
	}
	if !contains(config.IgnoreList, "rx_jumbo_ring_max") {
		if rxJumboRingMax, err := sysEthtool.RingRxJumboMax(); err == nil {
			e.RxJumboRingMax = rxJumboRingMax
		}
	}
	if !contains(config.IgnoreList, "priv_flags") {
		if privFlags, err := sysEthtool.PrivFlags(); err == nil {
			e.PrivFlags = privFlags
		}
	}
	return e, nil
}
