// +build windows

package system

func (f *DefFile) Mode() (string, error) {
	return "0000", nil // TODO implement
}

func (f *DefFile) Owner() (string, error) {
	return getUserForUid(1000)
}

func (f *DefFile) Group() (string, error) {
	return getGroupForGid(1000)
}
