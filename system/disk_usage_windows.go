package system

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func (u *DefDiskUsage) Exists() (bool, error) {
	if u.err != nil {
		if errN, ok := u.err.(windows.Errno); ok && errN == windows.ERROR_PATH_NOT_FOUND {
			return false, nil
		}
		return false, u.err
	}
	return true, nil
}

func (u *DefDiskUsage) Calculate() {
	var dummy uint64

	r1, _, err := windows.
		NewLazySystemDLL("kernel32.dll").
		NewProc("GetDiskFreeSpaceExW").
		Call(
			uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(u.path))),
			uintptr(unsafe.Pointer(&dummy)), // free bytes available to caller
			uintptr(unsafe.Pointer(&u.totalBytes)),
			uintptr(unsafe.Pointer(&u.freeBytes)))

	if r1 == 0 {
		// syscall errors out if r1 is zero. err is always not nil.
		u.err = err
		return
	}

	u.err = nil
}
