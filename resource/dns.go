package resource

import (
	"fmt"
	"strings"
	"time"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type DNS struct {
	Title       string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta        meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id          string  `json:"-" yaml:"-"`
	Resolve     string  `json:"resolve,omitempty" yaml:"resolve,omitempty"`
	Resolveable matcher `json:"resolveable,omitempty" yaml:"resolveable,omitempty"`
	Resolvable  matcher `json:"resolvable" yaml:"resolvable"`
	Addrs       matcher `json:"addrs,omitempty" yaml:"addrs,omitempty"`
	Timeout     int     `json:"timeout" yaml:"timeout"`
	Server      string  `json:"server,omitempty" yaml:"server,omitempty"`
	Skip        bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	DNSResourceKey  = "dns"
	DNSResourceName = "DNS"
)

func init() {
	registerResource(DNSResourceKey, &DNS{})
}

func (d *DNS) ID() string {
	if d.Resolve != "" && d.Resolve != d.id {
		return fmt.Sprintf("%s: %s", d.id, d.Resolve)
	}
	return d.id
}
func (d *DNS) SetID(id string)  { d.id = id }
func (d *DNS) SetSkip()         { d.Skip = true }
func (d *DNS) TypeKey() string  { return DNSResourceKey }
func (d *DNS) TypeName() string { return DNSResourceName }
func (d *DNS) GetTitle() string { return d.Title }
func (d *DNS) GetMeta() meta    { return d.Meta }
func (d *DNS) GetResolve() string {
	if d.Resolve != "" {
		return d.Resolve
	}
	return d.id
}

func (d *DNS) Validate(sys *system.System) []TestResult {
	skip := d.Skip
	if d.Timeout == 0 {
		d.Timeout = 500
	}

	sysDNS := sys.NewDNS(d.GetResolve(), sys, util.Config{Timeout: time.Duration(d.Timeout) * time.Millisecond, Server: d.Server})

	var results []TestResult
	// Backwards compatibility hack for now
	if d.Resolvable == nil {
		d.Resolvable = d.Resolveable
	}
	results = append(results, ValidateValue(d, "resolvable", d.Resolvable, sysDNS.Resolvable, skip))
	if shouldSkip(results) {
		skip = true
	}
	if d.Addrs != nil {
		results = append(results, ValidateValue(d, "addrs", d.Addrs, sysDNS.Addrs, skip))
	}
	return results
}

func NewDNS(sysDNS system.DNS, config util.Config) (*DNS, error) {
	var host string
	if sysDNS.Qtype() != "" {
		host = strings.Join([]string{sysDNS.Qtype(), sysDNS.Host()}, ":")
	} else {
		host = sysDNS.Host()
	}

	resolvable, err := sysDNS.Resolvable()
	server := sysDNS.Server()

	d := &DNS{
		id:         host,
		Resolvable: resolvable,
		Timeout:    config.TimeOutMilliSeconds(),
		Server:     server,
	}
	if !contains(config.IgnoreList, "addrs") {
		addrs, _ := sysDNS.Addrs()
		d.Addrs = addrs
	}
	return d, err
}
