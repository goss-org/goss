package resource

import "github.com/aelsabbahy/goss/system"

type DNS struct {
	Host        string   `json:"host"`
	Resolveable bool     `json:"resolveable"`
	Addrs       []string `json:"addrs,omitempty"`
	Timeout     int64    `json:"timeout"`
}

func (d *DNS) Validate(sys *system.System) []TestResult {
	sysDNS := sys.NewDNS(d.Host, sys)
	sysDNS.Timeout = d.Timeout

	var results []TestResult

	results = append(results, ValidateValue(d.Host, "resolveable", d.Resolveable, sysDNS.Resolveable))
	if !d.Resolveable {
		return results
	}
	results = append(results, ValidateValues(d.Host, "addrs", d.Addrs, sysDNS.Addrs))

	return results
}

func NewDNS(sysDNS system.DNS) *DNS {
	host := sysDNS.Host()
	addrs, _ := sysDNS.Addrs()
	resolveable, _ := sysDNS.Resolveable()
	return &DNS{
		Host:        host,
		Addrs:       addrs,
		Resolveable: resolveable.(bool),
		Timeout:     500,
	}
}
