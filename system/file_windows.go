// +build windows

package system

func (f *DefFile) Mode() (string, error) {
	return "windows", nil
}

func (f *DefFile) Owner() (string, error) {
	return "windows"
}

func (f *DefFile) Group() (string, error) {
	return "windows"
}
