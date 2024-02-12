package resource

import (
	"context"
	"fmt"
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
	"time"
)

type Mount struct {
	Title      string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta       meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	id         string  `json:"-" yaml:"-"`
	MountPoint string  `json:"mountpoint,omitempty" yaml:"mountpoint,omitempty"`
	Exists     matcher `json:"exists" yaml:"exists"`
	Opts       matcher `json:"opts,omitempty" yaml:"opts,omitempty"`
	VfsOpts    matcher `json:"vfs-opts,omitempty" yaml:"vfs-opts,omitempty"`
	Source     matcher `json:"source,omitempty" yaml:"source,omitempty"`
	Filesystem matcher `json:"filesystem,omitempty" yaml:"filesystem,omitempty"`
	Timeout    int     `json:"timeout" yaml:"timeout"`
	Skip       bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
	Usage      matcher `json:"usage,omitempty" yaml:"usage,omitempty"`
}

const (
	MountResourceKey  = "mount"
	MountResourceName = "Mount"
)

func init() {
	registerResource(MountResourceKey, &Mount{})
}

func (m *Mount) ID() string {
	if m.MountPoint != "" && m.MountPoint != m.id {
		return fmt.Sprintf("%s: %s", m.id, m.MountPoint)
	}
	return m.id
}
func (m *Mount) SetID(id string)  { m.id = id }
func (m *Mount) SetSkip()         { m.Skip = true }
func (m *Mount) TypeKey() string  { return MountResourceKey }
func (m *Mount) TypeName() string { return MountResourceName }

// FIXME: Can this be refactored?
func (m *Mount) GetTitle() string { return m.Title }
func (m *Mount) GetMeta() meta    { return m.Meta }
func (m *Mount) GetMountPoint() string {
	if m.MountPoint != "" {
		return m.MountPoint
	}
	return m.id
}

func (m *Mount) Validate(sys *system.System) []TestResult {
	ctx := context.WithValue(context.Background(), "id", m.ID())
	skip := m.Skip

	if m.Timeout == 0 {
		m.Timeout = 1000
	}

	sysMount := sys.NewMount(ctx, m.GetMountPoint(), sys, util.Config{Timeout: time.Duration(m.Timeout) * time.Millisecond})

	var results []TestResult
	results = append(results, ValidateValue(m, "exists", m.Exists, sysMount.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if m.Opts != nil {
		results = append(results, ValidateValue(m, "opts", m.Opts, sysMount.Opts, skip))
	}
	if m.VfsOpts != nil {
		results = append(results, ValidateValue(m, "vfs-opts", m.VfsOpts, sysMount.VfsOpts, skip))
	}
	if m.Source != nil {
		results = append(results, ValidateValue(m, "source", m.Source, sysMount.Source, skip))
	}
	if m.Filesystem != nil {
		results = append(results, ValidateValue(m, "filesystem", m.Filesystem, sysMount.Filesystem, skip))
	}
	if m.Usage != nil {
		results = append(results, ValidateValue(m, "usage", m.Usage, sysMount.Usage, skip))
	}
	return results
}

func NewMount(sysMount system.Mount, config util.Config) (*Mount, error) {
	mountPoint := sysMount.MountPoint()
	exists, _ := sysMount.Exists()
	m := &Mount{
		id:      mountPoint,
		Exists:  exists,
		Timeout: config.TimeOutMilliSeconds(),
	}
	if !contains(config.IgnoreList, "opts") {
		if opts, err := sysMount.Opts(); err == nil {
			m.Opts = opts
		}
	}
	if !contains(config.IgnoreList, "vfs-opts") {
		if vfsOpts, err := sysMount.VfsOpts(); err == nil {
			m.VfsOpts = vfsOpts
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
