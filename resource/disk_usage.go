package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type DiskUsage struct {
	Title              string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta               meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Path               string  `json:"-" yaml:"-"`
	Exists             matcher `json:"exists" yaml:"exists"`
	TotalBytes         matcher `json:"total_bytes" yaml:"total_bytes"`
	FreeBytes          matcher `json:"free_bytes" yaml:"free_bytes"`
	UtilizationPercent matcher `json:"utilization_percent" yaml:"utilization_percent"`
}

func (u *DiskUsage) ID() string      { return u.Path }
func (u *DiskUsage) SetID(id string) { u.Path = id }

func (u *DiskUsage) GetTitle() string { return u.Title }
func (u *DiskUsage) GetMeta() meta    { return u.Meta }

func (u *DiskUsage) Validate(sys *system.System) []TestResult {
	skip := false
	sysDU := sys.NewDiskUsage(u.Path, sys, util.Config{})
	sysDU.Calculate()

	var results []TestResult
	results = append(results, ValidateValue(u, "exists", u.Exists, sysDU.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if u.TotalBytes != nil {
		results = append(results, ValidateValue(u, "total_bytes", u.TotalBytes, sysDU.TotalBytes, skip))
	}
	if u.FreeBytes != nil {
		results = append(results, ValidateValue(u, "free_bytes", u.FreeBytes, sysDU.FreeBytes, skip))
	}
	if u.UtilizationPercent != nil {
		results = append(results, ValidateValue(u, "utilization_percent", u.UtilizationPercent, sysDU.UtilizationPercent, skip))
	}
	return results
}

func NewDiskUsage(sysDiskUsage system.DiskUsage, config util.Config) (*DiskUsage, error) {
	sysDiskUsage.Calculate()
	exists, _ := sysDiskUsage.Exists()
	u := &DiskUsage{
		Path:   sysDiskUsage.Path(),
		Exists: exists,
	}

	if !contains(config.IgnoreList, "total_bytes") {
		if totalBytes, err := sysDiskUsage.TotalBytes(); err != nil {
			u.TotalBytes = totalBytes
		}
	}
	if !contains(config.IgnoreList, "free_bytes") {
		if freeBytes, err := sysDiskUsage.FreeBytes(); err != nil {
			u.FreeBytes = freeBytes
		}
	}
	if !contains(config.IgnoreList, "utilization_percent") {
		if utilizationPercent, err := sysDiskUsage.UtilizationPercent(); err != nil {
			u.UtilizationPercent = utilizationPercent
		}
	}
	return u, nil
}
