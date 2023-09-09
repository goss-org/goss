package resource

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Interface struct {
	Title  string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta   meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id     string  `json:"-" yaml:"-"`
	Name   string  `json:"name,omitempty" yaml:"name,omitempty"`
	Exists matcher `json:"exists" yaml:"exists"`
	Addrs  matcher `json:"addrs,omitempty" yaml:"addrs,omitempty"`
	MTU    matcher `json:"mtu,omitempty" yaml:"mtu,omitempty"`
	Skip   bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	InterfaceResourceKey  = "interface"
	InterfaceResourceName = "Interface"
)

func init() {
	registerResource(InterfaceResourceKey, &Interface{})
}

func (i *Interface) ID() string {
	if i.Name != "" && i.Name != i.id {
		return fmt.Sprintf("%s: %s", i.id, i.Name)
	}
	return i.id
}
func (i *Interface) SetID(id string)  { i.id = id }
func (i *Interface) SetSkip()         { i.Skip = true }
func (i *Interface) TypeKey() string  { return InterfaceResourceKey }
func (i *Interface) TypeName() string { return InterfaceResourceName }

// FIXME: Can this be refactored?
func (i *Interface) GetTitle() string { return i.Title }
func (i *Interface) GetMeta() meta    { return i.Meta }
func (i *Interface) GetName() string {
	if i.Name != "" {
		return i.Name
	}
	return i.id
}

func (i *Interface) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), "id", i.ID())
	skip := i.Skip
	sysInterface := sys.NewInterface(ctx, i.GetName(), sys, util.Config{})

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
		id:     name,
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
