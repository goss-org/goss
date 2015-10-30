package resource

import "github.com/aelsabbahy/goss/system"

type Addr struct {
	Address   string `json:"-"`
	Reachable bool   `json:"reachable"`
	Timeout   int64  `json:"timeout"`
}

func (h *Addr) ID() string      { return h.Address }
func (h *Addr) SetID(id string) { h.Address = id }

func (h *Addr) Validate(sys *system.System) []TestResult {
	sysAddr := sys.NewAddr(h.Address, sys)
	sysAddr.SetTimeout(h.Timeout)

	var results []TestResult

	results = append(results, ValidateValue(h, "reachable", h.Reachable, sysAddr.Reachable))

	return results
}

func NewAddr(sysAddr system.Addr, ignoreList []string) *Addr {
	address := sysAddr.Address()
	reachable, _ := sysAddr.Reachable()
	return &Addr{
		Address:   address,
		Reachable: reachable.(bool),
		Timeout:   500,
	}
}
