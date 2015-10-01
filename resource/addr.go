package resource

import "github.com/aelsabbahy/goss/system"

type Addr struct {
	Address   string `json:"address"`
	Reachable bool   `json:"reachable"`
	Timeout   int64  `json:"timeout"`
}

func (h *Addr) Validate(sys *system.System) []TestResult {
	sysAddr := sys.NewAddr(h.Address, sys)
	sysAddr.Timeout = h.Timeout

	var results []TestResult

	results = append(results, ValidateValue(h.Address, "reachable", h.Reachable, sysAddr.Reachable))

	return results
}

func NewAddr(sysAddr system.Addr) *Addr {
	address := sysAddr.Address()
	reachable, _ := sysAddr.Reachable()
	return &Addr{
		Address:   address,
		Reachable: reachable.(bool),
		Timeout:   500,
	}
}
