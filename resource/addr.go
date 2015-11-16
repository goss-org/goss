package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Addr struct {
	Address   string `json:"-"`
	Reachable bool   `json:"reachable"`
	Timeout   int    `json:"timeout"`
}

func (h *Addr) ID() string      { return h.Address }
func (h *Addr) SetID(id string) { h.Address = id }

func (h *Addr) Validate(sys *system.System) []TestResult {
	sysAddr := sys.NewAddr(h.Address, sys, util.Config{Timeout: h.Timeout})

	var results []TestResult

	results = append(results, ValidateValue(h, "reachable", h.Reachable, sysAddr.Reachable))

	return results
}

func NewAddr(sysAddr system.Addr, config util.Config) (*Addr, error) {
	address := sysAddr.Address()
	reachable, err := sysAddr.Reachable()
	a := &Addr{
		Address:   address,
		Reachable: reachable.(bool),
		Timeout:   config.Timeout,
	}
	return a, err
}
