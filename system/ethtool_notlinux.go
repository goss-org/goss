//go:build !linux

package system

import (
	"context"

	"github.com/goss-org/goss/util"
)

// defEthtool is the stub used on platforms without ethtool(8). Every getter
// fails with ErrEthtoolUnsupported rather than reporting a bogus value, so a
// gossfile using the ethtool resource explains itself instead of silently
// passing.
type defEthtool struct {
	name string
}

func NewDefEthtool(_ context.Context, name string, _ *System, _ util.Config) Ethtool {
	return &defEthtool{name: name}
}

func (e *defEthtool) Name() string                    { return e.name }
func (e *defEthtool) Exists() (bool, error)           { return false, ErrEthtoolUnsupported }
func (e *defEthtool) Speed() (string, error)          { return "", ErrEthtoolUnsupported }
func (e *defEthtool) Duplex() (string, error)         { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RxChecksum() (string, error)     { return "", ErrEthtoolUnsupported }
func (e *defEthtool) TxChecksum() (string, error)     { return "", ErrEthtoolUnsupported }
func (e *defEthtool) AdaptiveRx() (string, error)     { return "", ErrEthtoolUnsupported }
func (e *defEthtool) AdaptiveTx() (string, error)     { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RxUsecs() (string, error)        { return "", ErrEthtoolUnsupported }
func (e *defEthtool) TxUsecs() (string, error)        { return "", ErrEthtoolUnsupported }
func (e *defEthtool) Driver() (string, error)         { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RingRx() (string, error)         { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RingTx() (string, error)         { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RingRxMini() (string, error)     { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RingRxJumbo() (string, error)    { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RingRxMax() (string, error)      { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RingTxMax() (string, error)      { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RingRxMiniMax() (string, error)  { return "", ErrEthtoolUnsupported }
func (e *defEthtool) RingRxJumboMax() (string, error) { return "", ErrEthtoolUnsupported }
func (e *defEthtool) PrivFlags() ([]string, error)    { return nil, ErrEthtoolUnsupported }

// sysfsDriver has nothing to read on platforms without sysfs.
func sysfsDriver(string) (string, bool) { return "", false }
