// +build windows

package system

func (f *DefFile) Mode() (string, error) {
	return "-1", nil // not applicable on Windows
}

func (f *DefFile) Owner() (string, error) {
	return "-1", nil // not applicable on Windows
}

func (f *DefFile) Group() (string, error) {
	return "-1", nil // not applicable on Windows
}
