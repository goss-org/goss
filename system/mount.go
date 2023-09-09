package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/goss-org/goss/util"
	"github.com/moby/sys/mountinfo"
	"github.com/samber/lo"
)

type Mount interface {
	MountPoint() string
	Exists(context.Context) (bool, error)
	Opts(context.Context) ([]string, error)
	VfsOpts(context.Context) ([]string, error)
	Source(context.Context) (string, error)
	Filesystem(context.Context) (string, error)
	Usage(context.Context) (int, error)
}

type DefMount struct {
	mountPoint string
	loaded     bool
	exists     bool
	mountInfo  *mountinfo.Info
	usage      int
	err        error
}

func NewDefMount(mountPoint string, system *System, config util.Config) Mount {
	return &DefMount{
		mountPoint: mountPoint,
	}
}

func (m *DefMount) setup() error {
	if m.loaded {
		return m.err
	}
	m.loaded = true

	mountInfo, err := getMount(m.mountPoint)
	if err != nil {
		m.exists = false
		m.err = err
		return m.err
	}
	m.mountInfo = mountInfo
	m.exists = true

	usage, err := getUsage(m.mountPoint)
	if err != nil {
		m.err = err
		return m.err
	}
	m.usage = usage

	return nil
}

func (m *DefMount) ID() string {
	return m.mountPoint
}

func (m *DefMount) MountPoint() string {
	return m.mountPoint
}

func (m *DefMount) Exists(ctx context.Context) (bool, error) {
	if err := m.setup(); err != nil {
		return false, nil
	}

	return m.exists, nil
}

func (m *DefMount) Opts(ctx context.Context) ([]string, error) {
	if err := m.setup(); err != nil {
		return nil, err
	}
	allOpts := splitMountInfo(m.mountInfo.Options)

	return lo.Uniq(allOpts), nil
}

func (m *DefMount) VfsOpts(ctx context.Context) ([]string, error) {
	if err := m.setup(); err != nil {
		return nil, err
	}
	opts := splitMountInfo(m.mountInfo.VFSOptions)
	return opts, nil
}

func (m *DefMount) Source(ctx context.Context) (string, error) {
	if err := m.setup(); err != nil {
		return "", err
	}

	return m.mountInfo.Source, nil
}

func (m *DefMount) Filesystem(ctx context.Context) (string, error) {
	if err := m.setup(); err != nil {
		return "", err
	}

	return m.mountInfo.FSType, nil
}

func (m *DefMount) Usage(ctx context.Context) (int, error) {
	if err := m.setup(); err != nil {
		return -1, err
	}

	return m.usage, nil
}

func getMount(mountpoint string) (*mountinfo.Info, error) {
	entries, err := mountinfo.GetMounts(mountinfo.SingleEntryFilter(mountpoint))
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("Mountpoint not found")
	}
	return entries[0], nil
}

func splitMountInfo(s string) []string {
	quoted := false
	return strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ','
	})
}
