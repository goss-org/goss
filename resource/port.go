package resource

import (
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Port struct {
	Title     string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta      meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Port      string  `json:"-" yaml:"-"`
	Listening matcher `json:"listening" yaml:"listening"`
	IP        matcher `json:"ip,omitempty" yaml:"ip,omitempty"`
	Skip      bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	PortResourceKey  = "port"
	PortResourceName = "Port"
)

func init() {
	registerResource(PortResourceKey, &Port{})
}

func (p *Port) ID() string       { return p.Port }
func (p *Port) SetID(id string)  { p.Port = id }
func (p *Port) SetSkip()         { p.Skip = true }
func (p *Port) TypeKey() string  { return PortResourceKey }
func (p *Port) TypeName() string { return PortResourceName }
func (p *Port) GetTitle() string { return p.Title }
func (p *Port) GetMeta() meta    { return p.Meta }

func (p *Port) Validate(sys *system.System) []TestResult {
	skip := p.Skip
	sysPort := sys.NewPort(p.Port, sys, util.Config{})

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
