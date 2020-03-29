package system

import (
	"github.com/aelsabbahy/goss/util"

	"github.com/jaypipes/ghw"

	"errors"
)

// BlockDevice interface of BlockDevice
type BlockDevice interface {
	Name() string
	Exists() (bool, error)
	IsRemovable() (bool, error)
	Size() (int, error)
	DriveType() (string, error)
	Controller() (string, error)
	PhysicalBlockSize() (int, error)
	Vendor() (string, error)
}

// DefBlockDevice struct of Blockdevice
type DefBlockDevice struct {
	removable         bool
	drivetype         string
	size              int
	vendor            string
	physicalblocksize int
	controller        string
	ghwBlockDevice    *ghw.Disk
	exists            bool
	name              string
	err               error
}

// NewDefBlockDevice create a DefBlockDevice with name
func NewDefBlockDevice(name string, systei *System, config util.Config) BlockDevice {
	return &DefBlockDevice{
		name: name,
	}
}

func (b *DefBlockDevice) setup() error {
	block, err := ghw.Block()
	if err != nil {
		b.err = err
		b.exists = false
		return b.err
	}
	for _, disk := range block.Disks {
		if disk.Name == b.name {
			b.ghwBlockDevice = disk
			b.exists = true
			return nil
		}
	}
	b.exists = false
	return errors.New("Can't find block device")
}

// ID get single test case ID
func (b *DefBlockDevice) ID() string {
	return b.name
}

// Name get single test case Name
func (b *DefBlockDevice) Name() string {
	return b.name
}

// Exists check if block device exists
func (b *DefBlockDevice) Exists() (bool, error) {
	if err := b.setup(); err != nil {
		return false, nil
	}

	return b.exists, nil
}

// IsRemovable check if block device is removeable
func (b *DefBlockDevice) IsRemovable() (bool, error) {
	if err := b.setup(); err != nil {
		return false, err
	}
	return b.ghwBlockDevice.IsRemovable, nil
}

// DriveType get drive type. This string will be "HDD", "FDD", "ODD", or "SSD"
func (b *DefBlockDevice) DriveType() (string, error) {
	if err := b.setup(); err != nil {
		return "", err
	}
	return b.ghwBlockDevice.DriveType.String(), nil
}

// Controller get controller of block device. This string will be "SCSI", "IDE", "virtio", "MMC", or "NVMe"
func (b *DefBlockDevice) Controller() (string, error) {
	if err := b.setup(); err != nil {
		return "", err
	}
	return b.ghwBlockDevice.StorageController.String(), nil
}

// PhysicalBlockSize get physical block size for disk, return bytes
func (b *DefBlockDevice) PhysicalBlockSize() (int, error) {
	if err := b.setup(); err != nil {
		return 0, err
	}
	return int(b.ghwBlockDevice.PhysicalBlockSizeBytes), nil
}

// Size get size for disk, return bytes
func (b *DefBlockDevice) Size() (int, error) {
	if err := b.setup(); err != nil {
		return 0, err
	}
	return int(b.ghwBlockDevice.SizeBytes), nil
}

// Vendor get vendor for disk
func (b *DefBlockDevice) Vendor() (string, error) {
	if err := b.setup(); err != nil {
		return "", err
	}
	return b.ghwBlockDevice.Vendor, nil
}
