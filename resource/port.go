package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Port struct {
	Desc      string  `json:"desc,omitempty" yaml:"desc,omitempty"`
	Port      string  `json:"-" yaml:"-"`
	Listening bool    `json:"listening" yaml:"listening"`
	IP        matcher `json:"ip,omitempty" yaml:"ip,omitempty"`
}

func (p *Port) ID() string      { return p.Port }
func (p *Port) SetID(id string) { p.Port = id }

func (p *Port) Validate(sys *system.System) []TestResult {
	sysPort := sys.NewPort(p.Port, sys, util.Config{})

	var results []TestResult

	results = append(results, ValidateValue(p, "listening", p.Listening, sysPort.Listening))

	if p.IP != nil {
		results = append(results, ValidateValue(p, "ip", p.IP, sysPort.IP))
	}

	return results
}

func NewPort(sysPort system.Port, config util.Config) (*Port, error) {
	port := sysPort.Port()
	listening, _ := sysPort.Listening()
	p := &Port{
		Port:      port,
		Listening: listening,
	}
	if !contains(config.IgnoreList, "ip") {
		if ip, err := sysPort.IP(); err == nil {
			p.IP = ip
		}
	}
	return p, nil
}
