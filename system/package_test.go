package system

import (
	"testing"
)

func TestIsSupportedPackageManager(t *testing.T) {
	if IsSupportedPackageManager("na") {
		t.Fatal("na should not be a valid package manager")
	}

	if !IsSupportedPackageManager("rpm") {
		t.Fatal("rpm should be a valid package manager")
	}
}
