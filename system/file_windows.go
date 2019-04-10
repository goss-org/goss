package system

import (
	"fmt"
	"os"
)

// Mode returns golang file mode. This method is cross-platform, but we are keeping unix-specific implementation
// in file_unix.go for backwards compatibility.
func (f *DefFile) Mode() (string, error) {
	if err := f.setup(); err != nil {
		return "", err
	}

	fi, err := os.Lstat(f.realPath)
	if err != nil {
		return "", err
	}

	mode := fmt.Sprintf("%04o", (fi.Mode() & 07777))
	return mode, nil
}

// TODO(ENG-1522): owner and group are not defined for windows in standard library, find out whether we can support it.
func (f *DefFile) Owner() (string, error) {
	return "", nil
}

func (f *DefFile) Group() (string, error) {
	return "", nil
}
