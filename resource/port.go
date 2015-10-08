package resource

import "github.com/aelsabbahy/goss/system"

type Port struct {
	Port      string `json:"port"`
	Listening bool   `json:"listening"`
	IP        string `json:"ip,omitempty"`
}

func (p *Port) ID() string { return p.Port }

func (p *Port) Validate(sys *system.System) []TestResult {
	sysPort := sys.NewPort(p.Port, sys)

	var results []TestResult

	results = append(results, ValidateValue(p.Port, "listening", p.Listening, sysPort.Listening))
	if !p.Listening {
		return results
	}
	results = append(results, ValidateValue(p.Port, "ip", p.IP, sysPort.IP))

	return results
}

func NewPort(sysPort system.Port) *Port {
	port := sysPort.Port()
	listening, _ := sysPort.Listening()
	ip, _ := sysPort.IP()
	return &Port{
		Port:      port,
		Listening: listening.(bool),
		IP:        ip.(string),
	}
}
