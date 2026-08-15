package resource

import (
	"context"
	"strings"
	"testing"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"

	"gopkg.in/yaml.v3"
)

// NewFile used to set Contents to an empty list, so every generated gossfile
// warned about asserting nothing. It must stay unset and marshal away.
func TestNewFileLeavesContentsUnset(t *testing.T) {
	sysFile := system.NewDefFile(context.Background(), "/etc/hostname", nil, util.Config{})
	f, err := NewFile(sysFile, util.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if f.Contents != nil {
		t.Errorf("NewFile Contents = %#v, want nil", f.Contents)
	}

	out, err := yaml.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "contents") {
		t.Errorf("marshalled file contains a contents key:\n%s", out)
	}
}
