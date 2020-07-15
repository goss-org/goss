package system

import (
	"fmt"
	"strings"

	"github.com/aelsabbahy/goss/util"
	"github.com/moby/sys/mountinfo"
	"github.com/thoas/go-funk"
)

type Mount interface {
	MountPoint() string
	Exists() (bool, error)
	Opts() ([]string, error)
	Source() (string, error)
	Filesystem() (string, error)
	Usage() (int, error)
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

func (m *DefMount) Exists() (bool, error) {
	if err := m.setup(); err != nil {
		return false, nil
	}

	return m.exists, nil
}

func (m *DefMount) Opts() ([]string, error) {
	if err := m.setup(); err != nil {
		return nil, err
	}
	allOpts := strings.Split(strings.Join([]string{m.mountInfo.Opts, m.mountInfo.VfsOpts}, ","), ",")

	return funk.UniqString(allOpts), nil
}

func (m *DefMount) Source() (string, error) {
	if err := m.setup(); err != nil {
		return "", err
	}

	return m.mountInfo.Source, nil
}

func (m *DefMount) Filesystem() (string, error) {
	if err := m.setup(); err != nil {
		return "", err
	}

	return m.mountInfo.Fstype, nil
}

func (m *DefMount) Usage() (int, error) {
	if err := m.setup(); err != nil {
		return -1, err
	}

	return m.usage, nil
}

func getMount(mountpoint string) (*mountinfo.Info, error) {
	entries, err := mountinfo.GetMounts(nil)
	if err != nil {
		return nil, err
	}

	// Search the table for the mountpoint
	for _, e := range entries {
		if e.Mountpoint == mountpoint {
			return e, nil
		}
	}
	return nil, fmt.Errorf("Mountpoint not found")
}
