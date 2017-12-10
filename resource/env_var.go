package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type EnvVar struct {
	Title string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta  meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Key   string  `json:"-" yaml:"-"`
	Value matcher `json:"value" yaml:"value"`
}

func (a *EnvVar) ID() string      { return a.Key }
func (a *EnvVar) SetID(id string) { a.Key = id }

// FIXME: Can this be refactored?
func (r *EnvVar) GetTitle() string { return r.Title }
func (r *EnvVar) GetMeta() meta    { return r.Meta }

func (a *EnvVar) Validate(sys *system.System) []TestResult {
	skip := false
	sysEnvVar := sys.NewEnvVar(a.Key, sys, util.Config{})

	var results []TestResult
	results = append(results, ValidateValue(a, "value", a.Value, sysEnvVar.Value, skip))
	return results
}

func NewEnvVar(sysEnvVar system.EnvVar, config util.Config) (*EnvVar, error) {
	key := sysEnvVar.Key()
	value, err := sysEnvVar.Value()
	a := &EnvVar{
		Key:   key,
		Value: value,
	}
	return a, err
}
