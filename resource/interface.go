package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Interface struct {
	Title  string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta   meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Name   string  `json:"-" yaml:"-"`
	Exists matcher `json:"exists" yaml:"exists"`
	Addrs  matcher `json:"addrs,omitempty" yaml:"addrs,omitempty"`
	MTU    matcher `json:"mtu,omitempty" yaml:"mtu,omitempty"`
}

func (i *Interface) ID() string      { return i.Name }
func (i *Interface) SetID(id string) { i.Name = id }

// FIXME: Can this be refactored?
func (i *Interface) GetTitle() string { return i.Title }
func (i *Interface) GetMeta() meta    { return i.Meta }

func (i *Interface) Validate(sys *system.System) []TestResult {
	skip := false
	sysInterface := sys.NewInterface(i.Name, sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(i, "exists", i.Exists, sysInterface.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if i.Addrs != nil {
		results = append(results, ValidateValue(i, "addrs", i.Addrs, sysInterface.Addrs, skip))
	}
	if i.MTU != nil {
		results = append(results, ValidateValue(i, "mtu", i.MTU, sysInterface.MTU, skip))
	}
	return results
}

func NewInterface(sysInterface system.Interface, config util.Config) (*Interface, error) {
	name := sysInterface.Name()
	exists, _ := sysInterface.Exists()
	i := &Interface{
		Name:   name,
		Exists: exists,
	}
	if !contains(config.IgnoreList, "addrs") {
		if addrs, err := sysInterface.Addrs(); err == nil {
			i.Addrs = addrs
		}
	}
	if !contains(config.IgnoreList, "mtu") {
		if mtu, err := sysInterface.MTU(); err == nil {
			i.MTU = mtu
		}
	}
	return i, nil
}
