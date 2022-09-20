package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type KernelParam struct {
	Title string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta  meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Key   string  `json:"-" yaml:"-"`
	Value matcher `json:"value" yaml:"value"`
}

const (
	KernelParamResourceKey  = "kernel-param"
	KernelParamResourceName = "KernelParam"
)

func init() {
	registerResource(KernelParamResourceKey, &KernelParam{})
}

func (a *KernelParam) ID() string      { return a.Key }
func (a *KernelParam) SetID(id string) { a.Key = id }

// FIXME: Can this be refactored?
func (r *KernelParam) GetTitle() string { return r.Title }
func (r *KernelParam) GetMeta() meta    { return r.Meta }

func (a *KernelParam) Validate(sys *system.System, skipTypes []string) []TestResult {
	skip := util.IsValueInList(KernelParamResourceKey, skipTypes)
	sysKernelParam := sys.NewKernelParam(a.Key, sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(a, "value", a.Value, sysKernelParam.Value, skip))
	return results
}

func NewKernelParam(sysKernelParam system.KernelParam, config util.Config) (*KernelParam, error) {
	key := sysKernelParam.Key()
	value, err := sysKernelParam.Value()
	a := &KernelParam{
		Key:   key,
		Value: value,
	}
	return a, err
}
