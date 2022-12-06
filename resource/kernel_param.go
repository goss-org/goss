package resource

import (
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type KernelParam struct {
	Title string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta  meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
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

func (a *KernelParam) ID() string       { return a.Key }
func (a *KernelParam) SetID(id string)  { a.Key = id }
func (a *KernelParam) SetSkip()         { a.Skip = true }
func (a *KernelParam) TypeKey() string  { return KernelParamResourceKey }
func (a *KernelParam) TypeName() string { return KernelParamResourceName }

// FIXME: Can this be refactored?
func (a *KernelParam) GetTitle() string { return a.Title }
func (a *KernelParam) GetMeta() meta    { return a.Meta }

func (a *KernelParam) Validate(sys *system.System) []TestResult {
	skip := a.Skip
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
