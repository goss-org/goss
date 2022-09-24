package resource

import (
	"fmt"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Port struct {
	Title     string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta      meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id        string  `json:"-" yaml:"-"`
	Port      string  `json:"port,omitempty" yaml:"port,omitempty"`
	Listening matcher `json:"listening" yaml:"listening"`
	IP        matcher `json:"ip,omitempty" yaml:"ip,omitempty"`
	Skip      bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

func (p *Port) ID() string {
	if p.Port != "" && p.Port != p.id {
		return fmt.Sprintf("%s: %s", p.id, p.Port)
	}
	return p.id
}
func (p *Port) SetID(id string) { p.id = id }

func (p *Port) GetTitle() string { return p.Title }
func (p *Port) GetMeta() meta    { return p.Meta }
func (p *Port) GetPort() string {
	if p.Port != "" {
		return p.Port
	}
	return p.id
}

func (p *Port) Validate(sys *system.System) []TestResult {
	skip := false
	sysPort := sys.NewPort(p.GetPort(), sys, util.Config{})

	if p.Skip {
		skip = true
	}

	var results []TestResult
	results = append(results, ValidateValue(p, "listening", p.Listening, sysPort.Listening, skip))
	if shouldSkip(results) {
		skip = true
	}
	if p.IP != nil {
		results = append(results, ValidateValue(p, "ip", p.IP, sysPort.IP, skip))
	}
	return results
}

func NewPort(sysPort system.Port, config util.Config) (*Port, error) {
	port := sysPort.Port()
	listening, _ := sysPort.Listening()
	p := &Port{
		id:        port,
		Listening: listening,
	}
	if !contains(config.IgnoreList, "ip") {
		if ip, err := sysPort.IP(); err == nil {
			p.IP = ip
		}
	}
	return p, nil
}
