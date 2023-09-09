package resource

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
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

const (
	PortResourceKey  = "port"
	PortResourceName = "Port"
)

func init() {
	registerResource(PortResourceKey, &Port{})
}

func (p *Port) ID() string {
	if p.Port != "" && p.Port != p.id {
		return fmt.Sprintf("%s: %s", p.id, p.Port)
	}
	return p.id
}
func (p *Port) SetID(id string)  { p.id = id }
func (p *Port) SetSkip()         { p.Skip = true }
func (p *Port) TypeKey() string  { return PortResourceKey }
func (p *Port) TypeName() string { return PortResourceName }
func (p *Port) GetTitle() string { return p.Title }
func (p *Port) GetMeta() meta    { return p.Meta }
func (p *Port) GetPort() string {
	if p.Port != "" {
		return p.Port
	}
	return p.id
}

func (p *Port) Validate(sys *system.System) []TestResult {
	ctx := context.Background()
	skip := p.Skip
	sysPort := sys.NewPort(p.GetPort(), sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(ctx, p, "listening", p.Listening, sysPort.Listening, skip))
	if shouldSkip(results) {
		skip = true
	}
	if p.IP != nil {
		results = append(results, ValidateValue(ctx, p, "ip", p.IP, sysPort.IP, skip))
	}
	return results
}

func NewPort(sysPort system.Port, config util.Config) (*Port, error) {
	ctx := context.Background()
	port := sysPort.Port()
	listening, _ := sysPort.Listening(ctx)
	p := &Port{
		id:        port,
		Listening: listening,
	}
	if !contains(config.IgnoreList, "ip") {
		if ip, err := sysPort.IP(ctx); err == nil {
			p.IP = ip
		}
	}
	return p, nil
}
