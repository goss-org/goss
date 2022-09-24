package resource

import (
	"fmt"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type KernelParam struct {
	Title string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta  meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id    string  `json:"-" yaml:"-"`
	Name  string  `json:"name,omitempty" yaml:"name,omitempty"`
	Key   string  `json:"-" yaml:"-"`
	Value matcher `json:"value" yaml:"value"`
}

func (k *KernelParam) ID() string {
	if k.Name != "" && k.Name != k.id {
		return fmt.Sprintf("%s: %s", k.id, k.Name)
	}
	return k.id
}
func (a *KernelParam) SetID(id string) { a.id = id }

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
	skip := false
	sysKernelParam := sys.NewKernelParam(k.GetName(), sys, util.Config{})

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
