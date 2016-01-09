package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Port struct {
	Port      string   `json:"-"`
	Listening bool     `json:"listening"`
	IP        []string `json:"ip,omitempty"`
}

func (p *Port) ID() string      { return p.Port }
func (p *Port) SetID(id string) { p.Port = id }

func (p *Port) Validate(sys *system.System) []TestResult {
	sysPort := sys.NewPort(p.Port, sys, util.Config{})

	var results []TestResult

	results = append(results, ValidateValue(p, "listening", p.Listening, sysPort.Listening))

	if len(p.IP) > 0 {
		results = append(results, ValidateValues(p, "ip", p.IP, sysPort.IP))
	}

	return results
}

func NewPort(sysPort system.Port, config util.Config) (*Port, error) {
	port := sysPort.Port()
	listening, _ := sysPort.Listening()
	p := &Port{
		Port:      port,
		Listening: listening.(bool),
	}
	if !contains(config.IgnoreList, "ip") {
		ip, _ := sysPort.IP()
		p.IP = ip
	}
	return p, nil
}
