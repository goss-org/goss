package resource

import "github.com/aelsabbahy/goss/system"

type Port struct {
	Port      string `json:"-"`
	Listening bool   `json:"listening"`
	IP        string `json:"ip,omitempty"`
}

func (p *Port) ID() string      { return p.Port }
func (p *Port) SetID(id string) { p.Port = id }

func (p *Port) Validate(sys *system.System) []TestResult {
	sysPort := sys.NewPort(p.Port, sys)

	var results []TestResult

	results = append(results, ValidateValue(p, "listening", p.Listening, sysPort.Listening))

	if p.IP != "" {
		results = append(results, ValidateValue(p, "ip", p.IP, sysPort.IP))
	}

	return results
}

func NewPort(sysPort system.Port, ignoreList []string) *Port {
	port := sysPort.Port()
	listening, _ := sysPort.Listening()
	p := &Port{
		Port:      port,
		Listening: listening.(bool),
	}
	if !contains(ignoreList, "ip") {
		ip, _ := sysPort.IP()
		p.IP = ip.(string)
	}
	return p
}
