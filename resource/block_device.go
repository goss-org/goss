package resource

import (
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

// BlockDevice define the basic infomation fo block device
type BlockDevice struct {
	Title             string  `json:"title,omitempty" yaml:"title,omitempty"`
	Meta              meta    `json:"meta,omitempty" yaml:"meta,omitempty"`
	Name              string  `json:"-" yaml:"-"`
	Exists            bool    `json:"exists" yaml:"exists"`
	IsRemovable       bool    `json:"isremovable,omitempty" yaml:"isremovable,omitempty"`
	DriveType         string  `json:"drivetype,omitempty" yaml:"drivetype,omitempty"`
	Controller        string  `json:"controller,omitempty" yaml:"controller,omitempty"`
	PhysicalBlockSize matcher `json:"physicalblocksize,omitempty" yaml:"physicalblocksize,omitempty"`
	Size              matcher `json:"size,omitempty" yaml:"size,omitempty"`
	Vendor            string  `json:"vendor,omitempty" yaml:"vendor,omitempty"`
	Skip              bool    `json:"skip,omitempty" yaml:"skip,omitempty"`
}

// ID get block device's ID
func (b *BlockDevice) ID() string { return b.Name }

// SetID set block device's ID with block device's name
func (b *BlockDevice) SetID(id string) { b.Name = id }

// GetTitle Get title of block device
func (b *BlockDevice) GetTitle() string { return b.Title }

// GetMeta get meta of block device
func (b *BlockDevice) GetMeta() meta { return b.Meta }

// Validate validate a block device
func (b *BlockDevice) Validate(sys *system.System) []TestResult {
	skip := false
	sysBlockDevice := sys.NewBlockDevice(b.Name, sys, util.Config{})

	if b.Skip {
		skip = true
	}

	var results []TestResult

	results = append(results, ValidateValue(b, "exists", b.Exists, sysBlockDevice.Exists, skip))
	if shouldSkip(results) {
		skip = true
	}
	if b.IsRemovable != false {
		results = append(results, ValidateValue(b, "isremovable", b.IsRemovable, sysBlockDevice.IsRemovable, skip))
	}
	if b.DriveType != "" {
		results = append(results, ValidateValue(b, "drivetype", b.DriveType, sysBlockDevice.DriveType, skip))
	}
	if b.Controller != "" {
		results = append(results, ValidateValue(b, "controller", b.Controller, sysBlockDevice.Controller, skip))
	}
	if b.PhysicalBlockSize != nil {
		results = append(results, ValidateValue(b, "physicalblocksize", b.PhysicalBlockSize, sysBlockDevice.PhysicalBlockSize, skip))
	}
	if b.Size != nil {
		results = append(results, ValidateValue(b, "size", b.Size, sysBlockDevice.Size, skip))
	}
	if b.Vendor != "" {
		results = append(results, ValidateValue(b, "vendor", b.Vendor, sysBlockDevice.Vendor, skip))
	}
	return results
}

// NewBlockDevice get new block device struct
func NewBlockDevice(sysBlockDevice system.BlockDevice, config util.Config) (*BlockDevice, error) {
	name := sysBlockDevice.Name()
	exists, _ := sysBlockDevice.Exists()
	b := &BlockDevice{
		Name:   name,
		Exists: exists,
	}
	if !contains(config.IgnoreList, "isremovable") {
		if isremovable, err := sysBlockDevice.IsRemovable(); err == nil {
			b.IsRemovable = isremovable
		}
	}
	if !contains(config.IgnoreList, "drivetype") {
		if drivetype, err := sysBlockDevice.DriveType(); err == nil {
			b.DriveType = drivetype
		}
	}
	if !contains(config.IgnoreList, "controller") {
		if Controller, err := sysBlockDevice.Controller(); err == nil {
			b.Controller = Controller
		}
	}
	if !contains(config.IgnoreList, "physicalblocksize") {
		if physicalblocksize, err := sysBlockDevice.PhysicalBlockSize(); err == nil {
			b.PhysicalBlockSize = physicalblocksize
		}
	}
	if !contains(config.IgnoreList, "size") {
		if size, err := sysBlockDevice.Size(); err == nil {
			b.Size = size
		}
	}
	if !contains(config.IgnoreList, "vendor") {
		if vendor, err := sysBlockDevice.Vendor(); err == nil {
			b.Vendor = vendor
		}
	}
	return b, nil
}
