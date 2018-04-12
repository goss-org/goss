package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Addr struct {
	Title        string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta         meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Address      string  `json:"-" yaml:"-"`
	LocalAddress string  `json:"localaddress" yaml:"localaddress,omitempty"`
	Reachable    matcher `json:"reachable" yaml:"reachable"`
	Timeout      int     `json:"timeout" yaml:"timeout"`
}

func (a *Addr) ID() string      { return a.Address }
func (a *Addr) SetID(id string) { a.Address = id }

// FIXME: Can this be refactored?
func (r *Addr) GetTitle() string { return r.Title }
func (r *Addr) GetMeta() meta    { return r.Meta }

func (a *Addr) Validate(sys *system.System) []TestResult {
	skip := false
	if a.Timeout == 0 {
		a.Timeout = 500
	}

	sysAddr := sys.NewAddr(a.Address, sys, util.Config{Timeout: a.Timeout, LocalAddress: a.LocalAddress})

	var results []TestResult
	results = append(results, ValidateValue(a, "reachable", a.Reachable, sysAddr.Reachable, skip))
	return results
}

func NewAddr(sysAddr system.Addr, config util.Config) (*Addr, error) {
	address := sysAddr.Address()
	reachable, err := sysAddr.Reachable()
	a := &Addr{
		Address:      address,
		Reachable:    reachable,
		Timeout:      config.Timeout,
		LocalAddress: config.LocalAddress,
	}
	return a, err
}
