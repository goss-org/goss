package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type DNS struct {
	Host        string  `json:"-"`
	Resolveable bool    `json:"resolveable"`
	Addrs       matcher `json:"addrs,omitempty"`
	Timeout     int     `json:"timeout"`
}

func (d *DNS) ID() string      { return d.Host }
func (d *DNS) SetID(id string) { d.Host = id }

func (d *DNS) Validate(sys *system.System) []TestResult {
	if d.Timeout == 0 {
		d.Timeout = 500
	}
	sysDNS := sys.NewDNS(d.Host, sys, util.Config{Timeout: d.Timeout})

	var results []TestResult

	results = append(results, ValidateValue(d, "resolveable", d.Resolveable, sysDNS.Resolveable))

	if d.Addrs != nil {
		results = append(results, ValidateValue(d, "addrs", d.Addrs, sysDNS.Addrs))
	}

	return results
}

func NewDNS(sysDNS system.DNS, config util.Config) (*DNS, error) {
	host := sysDNS.Host()
	resolveable, err := sysDNS.Resolveable()
	d := &DNS{
		Host:        host,
		Resolveable: resolveable,
		Timeout:     config.Timeout,
	}
	if !contains(config.IgnoreList, "addrs") {
		addrs, _ := sysDNS.Addrs()
		d.Addrs = addrs
	}
	return d, err
}
