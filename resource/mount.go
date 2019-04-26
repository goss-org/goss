package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type Mount struct {
	Title      string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	MountPoint string  `json:"-" yaml:"-"`
	Exists     matcher `json:"exists" yaml:"exists"`
	Opts       matcher `json:"opts,omitempty" yaml:"opts,omitempty"`
	Source     matcher `json:"source,omitempty" yaml:"source,omitempty"`
	Filesystem matcher `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
    Skip       bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

func (m *Mount) ID() string      { return m.MountPoint }
func (m *Mount) SetID(id string) { m.MountPoint = id }

// FIXME: Can this be refactored?
func (m *Mount) GetTitle() string { return m.Title }
func (m *Mount) GetMeta() meta    { return m.Meta }

func (m *Mount) Validate(sys *system.System) []TestResult {
	skip := false
	sysMount := sys.NewMount(m.MountPoint, sys, util.Config{})

    if m.Skip { skip = true }

	var results []TestResult
	results = append(results, ValidateValue(m, "exists", m.Exists, sysMount.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if m.Opts != nil {
		results = append(results, ValidateValue(m, "opts", m.Opts, sysMount.Opts, skip))
	}
	if m.Source != nil {
		results = append(results, ValidateValue(m, "source", m.Source, sysMount.Source, skip))
	}
	if m.Filesystem != nil {
		results = append(results, ValidateValue(m, "filesystem", m.Filesystem, sysMount.Filesystem, skip))
	}
	return results
}

func NewMount(sysMount system.Mount, config util.Config) (*Mount, error) {
	mountPoint := sysMount.MountPoint()
	exists, _ := sysMount.Exists()
	m := &Mount{
		MountPoint: mountPoint,
		Exists:     exists,
	}
	if !contains(config.IgnoreList, "opts") {
		if opts, err := sysMount.Opts(); err == nil {
			m.Opts = opts
		}
	}
	if !contains(config.IgnoreList, "source") {
		if source, err := sysMount.Source(); err == nil {
			m.Source = source
		}
	}
	if !contains(config.IgnoreList, "filesystem") {
		if filesystem, err := sysMount.Filesystem(); err == nil {
			m.Filesystem = filesystem
		}
	}
	return m, nil
}
