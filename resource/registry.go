package resource

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type Registry struct {
	Title  string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta   meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id     string  `json:"-" yaml:"-"`
	Name   string  `json:"name,omitempty" yaml:"name,omitempty"`
	Exists matcher `json:"exists" yaml:"exists"`
	Value  matcher `json:"value,omitempty" yaml:"value,omitempty"`
	Type   matcher `json:"type,omitempty" yaml:"type,omitempty"`
	Skip   bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	RegistryResourceKey  = "registry"
	RegistryResourceName = "Registry"
)

func init() {
	registerResource(RegistryResourceKey, &Registry{})
}

func (r *Registry) ID() string {
	if r.Name != "" && r.Name != r.id {
		return fmt.Sprintf("%s: %s", r.id, r.Name)
	}
	return r.id
}

func (r *Registry) SetID(id string)  { r.id = id }
func (r *Registry) SetSkip()         { r.Skip = true }
func (r *Registry) TypeKey() string  { return RegistryResourceKey }
func (r *Registry) TypeName() string { return RegistryResourceName }
func (r *Registry) GetTitle() string { return r.Title }
func (r *Registry) GetMeta() meta    { return r.Meta }
func (r *Registry) GetName() string {
	if r.Name != "" {
		return r.Name
	}
	return r.id
}

func (r *Registry) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), idKey{}, r.ID())
	skip := r.Skip
	sysRegistry := sys.NewRegistry(ctx, r.GetName(), sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(r, "exists", r.Exists, sysRegistry.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if r.Value != nil {
		results = append(results, ValidateValue(r, "value", r.Value, sysRegistry.Value, skip))
	}
	if r.Type != nil {
		results = append(results, ValidateValue(r, "type", r.Type, sysRegistry.Type, skip))
	}
	return results
}

func NewRegistry(sysRegistry system.Registry, config util.Config) (*Registry, error) {
	key := sysRegistry.Key()
	exists, _ := sysRegistry.Exists()
	if !exists {
		return &Registry{
			id:     key,
			Exists: exists,
		}, nil
	}
	value, err := sysRegistry.Value()
	if err != nil {
		return nil, err
	}
	regType, err := sysRegistry.Type()
	if err != nil {
		return nil, err
	}
	return &Registry{
		id:     key,
		Exists: exists,
		Value:  value,
		Type:   regType,
	}, nil
}
