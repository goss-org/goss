package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type ReverseDNS struct {
	Title       string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta        meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Addr        string  `json:"-" yaml:"-"`
	Resolveable matcher `json:"resolveable" yaml:"resolveable"`
	Hosts       matcher `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Timeout     int     `json:"timeout" yaml:"timeout"`
}

func (d *ReverseDNS) ID() string      { return d.Addr }
func (d *ReverseDNS) SetID(id string) { d.Addr = id }

func (d *ReverseDNS) GetTitle() string { return d.Title }
func (d *ReverseDNS) GetMeta() meta    { return d.Meta }

func (d *ReverseDNS) Validate(sys *system.System) []TestResult {
	skip := false
	if d.Timeout == 0 {
		d.Timeout = 500
	}
	sysReverseDNS := sys.NewReverseDNS(d.Addr, sys, util.Config{Timeout: d.Timeout})

	var results []TestResult
	results = append(results, ValidateValue(d, "resolveable", d.Resolveable, sysReverseDNS.Resolveable, skip))
	if shouldSkip(results) {
		skip = true
	}
	if d.Hosts != nil {
		results = append(results, ValidateValue(d, "hosts", d.Hosts, sysReverseDNS.Hosts, skip))
	}
	return results
}

func NewReverseDNS(sysReverseDNS system.ReverseDNS, config util.Config) (*ReverseDNS, error) {
	addr := sysReverseDNS.Addr()
	resolveable, err := sysReverseDNS.Resolveable()
	d := &ReverseDNS{
		Addr:        addr,
		Resolveable: resolveable,
		Timeout:     config.Timeout,
	}
	if !contains(config.IgnoreList, "hosts") {
		hosts, _ := sysReverseDNS.Hosts()
		d.Hosts = hosts
	}
	return d, err
}
