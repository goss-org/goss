package resource

import (
	"context"
	"fmt"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type KernelParam struct {
	Title string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta  meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id    string  `json:"-" yaml:"-"`
	Name  string  `json:"name,omitempty" yaml:"name,omitempty"`
	Key   string  `json:"-" yaml:"-"`
	Value matcher `json:"value" yaml:"value"`
	Skip  bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

const (
	KernelParamResourceKey  = "kernel-param"
	KernelParamResourceName = "KernelParam"
)

func init() {
	registerResource(KernelParamResourceKey, &KernelParam{})
}

func (k *KernelParam) ID() string {
	if k.Name != "" && k.Name != k.id {
		return fmt.Sprintf("%s: %s", k.id, k.Name)
	}
	return k.id
}
func (a *KernelParam) SetID(id string) { a.id = id }

func (a *KernelParam) SetSkip()         { a.Skip = true }
func (a *KernelParam) TypeKey() string  { return KernelParamResourceKey }
func (a *KernelParam) TypeName() string { return KernelParamResourceName }

// FIXME: Can this be refactored?
func (k *KernelParam) GetTitle() string { return k.Title }
func (k *KernelParam) GetMeta() meta    { return k.Meta }
func (k *KernelParam) GetName() string {
	if k.Name != "" {
		return k.Name
	}
	return k.id
}

func (k *KernelParam) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), idKey{}, k.ID())
	skip := k.Skip
	sysKernelParam := sys.NewKernelParam(ctx, k.GetName(), sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(k, "value", k.Value, sysKernelParam.Value, skip))
	return results
}

func NewKernelParam(sysKernelParam system.KernelParam, config util.Config) (*KernelParam, error) {
	key := sysKernelParam.Key()
	value, err := sysKernelParam.Value()
	a := &KernelParam{
		id:    key,
		Value: value,
	}
	return a, err
}
