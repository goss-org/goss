package resource

import (
	"io"
	"strings"
	"testing"

	"github.com/goss-org/goss/util"
	"go.yaml.in/yaml/v3"
)

// fakeSysHTTP is a minimal system.HTTP implementation used to drive NewHTTP
// in tests without making a real network call.
type fakeSysHTTP struct{}

func (f *fakeSysHTTP) HTTP() string                { return "https://example.com" }
func (f *fakeSysHTTP) Status() (int, error)        { return 200, nil }
func (f *fakeSysHTTP) Headers() (io.Reader, error) { return nil, nil }
func (f *fakeSysHTTP) Body() (io.Reader, error)    { return nil, nil }
func (f *fakeSysHTTP) Exists() (bool, error)       { return true, nil }
func (f *fakeSysHTTP) SetAllowInsecure(bool)       {}
func (f *fakeSysHTTP) SetNoFollowRedirects(bool)   {}
func (f *fakeSysHTTP) Close() error                { return nil }

// TestNewHTTPLeavesBodyUnset pins the fix for `goss add http` writing
// `body: []` into generated gossfiles: an empty list is falsy for
// isSetWarnEmpty/ValidateValue, so goss validate would warn on goss's own
// generated output. NewHTTP must leave Body nil (matching the field's
// yaml:"body,omitempty" tag) rather than initializing it to []string{}.
func TestNewHTTPLeavesBodyUnset(t *testing.T) {
	h, err := NewHTTP(&fakeSysHTTP{}, util.Config{})
	if err != nil {
		t.Fatalf("NewHTTP returned error: %v", err)
	}
	if h.Body != nil {
		t.Errorf("NewHTTP().Body = %#v, want nil", h.Body)
	}

	out, err := yaml.Marshal(h)
	if err != nil {
		t.Fatalf("yaml.Marshal returned error: %v", err)
	}
	if strings.Contains(string(out), "body:") {
		t.Errorf("marshalled yaml contains a body key, want it omitted:\n%s", out)
	}
}
