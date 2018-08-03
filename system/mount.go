package system

import (
	"fmt"
	"strings"

	"github.com/aelsabbahy/goss/util"
	"github.com/docker/docker/pkg/mount"
)

type Mount interface {
	MountPoint() string
	Exists() (bool, error)
	Opts() ([]string, error)
	Source() (string, error)
	Filesystem() (string, error)
}

type DefMount struct {
	mountPoint string
	loaded     bool
	exists     bool
	mountInfo  *mount.Info
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

	return strings.Split(m.mountInfo.Opts, ","), nil
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

func getMount(mountpoint string) (*mount.Info, error) {
	entries, err := mount.GetMounts(nil)
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
