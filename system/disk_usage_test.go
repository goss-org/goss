package system

import (
	"testing"

	"github.com/aelsabbahy/goss/util"
)

func TestDiskUsageOK(t *testing.T) {
	u := NewDefDiskUsage("/", nil, util.Config{})
	u.Calculate()

	ex, err := u.Exists()
	if err != nil {
		t.Fatal(err)
	}
	if !ex {
		t.Fatal("/ does not exist")
	}

	total, err := u.TotalBytes()
	if err != nil {
		t.Fatal(err)
	}
	free, err := u.TotalBytes()
	if err != nil {
		t.Fatal(err)
	}
	if total < free {
		t.Fatalf("total(%v) is less than free(%v)", total, free)
	}
	if total <= 0 {
		t.Fatalf("total(%v) <= 0", total)
	}
}

func TestDiskUsageInvalid(t *testing.T) {
	u := NewDefDiskUsage("INVALID DIRECTORY", nil, util.Config{})
	u.Calculate()

	// Exist should return false and not fail.
	exist, err := u.Exists()
	if exist {
		t.Fatal("'INVALID DIRECTORY' existence check succeeded (and it should not have)")
	}
	if err != nil {
		t.Fatal(err)
	}

	// But totalBytes should error out.
	if _, err = u.TotalBytes(); err == nil {
		t.Fatal("Stat should fail on invalid directory")
	}
}

func TestUtilization(t *testing.T) {
	u := &DefDiskUsage{
		totalBytes: 0,
		freeBytes:  0,
		err:        nil,
	}

	util, err := u.UtilizationPercent()
	if err != nil {
		t.Fatal(err)
	}
	if util != 100 {
		t.Fatal("DiskUsage.Utilization should report 100% if disk has no space")
	}

	u.totalBytes = 100
	u.freeBytes = 80
	util, err = u.UtilizationPercent()
	if util != 20 {
		t.Fatalf("Utilization incorrect, got: %v, want: 20", util)
	}
}
