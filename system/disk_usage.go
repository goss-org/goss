package system

import "github.com/aelsabbahy/goss/util"

type DiskUsage interface {
	Exists() (bool, error)
	Path() string
	Calculate()
	TotalBytes() (uint64, error)
	FreeBytes() (uint64, error)
	UtilizationPercent() (int, error)
}

type DefDiskUsage struct {
	path        string
	totalBytes  uint64
	freeBytes   uint64
	err         error
	initialized bool
}

func NewDefDiskUsage(path string, system *System, config util.Config) DiskUsage {
	return &DefDiskUsage{
		path: path,
	}
}

func (u *DefDiskUsage) Path() string {
	return u.path
}

func (u *DefDiskUsage) TotalBytes() (uint64, error) {
	return u.totalBytes, u.err
}

func (u *DefDiskUsage) FreeBytes() (uint64, error) {
	return u.freeBytes, u.err
}

func (u *DefDiskUsage) UtilizationPercent() (int, error) {
	if u.err != nil {
		return 0, u.err
	}
	if u.totalBytes == 0 {
		// If totalBytes is 0, set utilization to 100%. This protects us from division by zero.
		return 100, nil
	}
	return 100 - int(u.freeBytes*100/u.totalBytes), nil
}
