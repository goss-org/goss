package resource

import "github.com/aelsabbahy/goss/system"

type DNS struct {
	Host        string   `json:"-"`
	Resolveable bool     `json:"resolveable"`
	Addrs       []string `json:"addrs,omitempty"`
	Timeout     int64    `json:"timeout"`
}

func (d *DNS) ID() string      { return d.Host }
func (d *DNS) SetID(id string) { d.Host = id }

func (d *DNS) Validate(sys *system.System) []TestResult {
	sysDNS := sys.NewDNS(d.Host, sys)
	sysDNS.SetTimeout(d.Timeout)

	var results []TestResult

	results = append(results, ValidateValue(d, "resolveable", d.Resolveable, sysDNS.Resolveable))

	if len(d.Addrs) > 0 {
		results = append(results, ValidateValues(d, "addrs", d.Addrs, sysDNS.Addrs))
	}

	return results
}

func NewDNS(sysDNS system.DNS, ignoreList []string) *DNS {
	host := sysDNS.Host()
	resolveable, _ := sysDNS.Resolveable()
	d := &DNS{
		Host:        host,
		Resolveable: resolveable.(bool),
		Timeout:     500,
	}
	if !contains(ignoreList, "addrs") {
		addrs, _ := sysDNS.Addrs()
		d.Addrs = addrs
	}
	return d
}
