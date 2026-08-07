package resource

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr runs fn with os.Stderr redirected and returns what it wrote.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Stderr = orig
		_ = w.Close()
		_ = r.Close()
	}()

	os.Stderr = w
	fn()
	_ = w.Close()
	os.Stderr = orig

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestIsSet(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want bool
	}{
		{"nil is unset", nil, false},
		{"empty list is unset", []any{}, false},
		{"list with one entry is set", []any{"foo"}, true},
		{"empty string is set", "", true},
		{"string is set", "foo", true},
		{"false is set", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isSet(c.in); got != c.want {
				t.Errorf("isSet(%#v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

// TestIsSetWarnEmpty pins both halves of the contract: the warning fires only
// for a list written out as empty, and the set/unset answer is byte-for-byte
// what isSet already returned, so no matcher changes outcome.
func TestIsSetWarnEmpty(t *testing.T) {
	cases := []struct {
		name     string
		in       any
		wantWarn bool
	}{
		{"empty list warns", []any{}, true},
		{"list with one entry does not warn", []any{"foo"}, false},
		{"nil does not warn", nil, false},
		{"empty string does not warn", "", false},
		{"string does not warn", "foo", false},
		{"empty map does not warn", map[string]any{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got bool
			out := captureStderr(t, func() {
				got = isSetWarnEmpty(c.in, "/some/id: file.contents")
			})

			if want := isSet(c.in); got != want {
				t.Errorf("isSetWarnEmpty(%#v) = %v, want isSet's %v", c.in, got, want)
			}

			warned := strings.Contains(out, "WARNING:")
			if warned != c.wantWarn {
				t.Errorf("isSetWarnEmpty(%#v) warned = %v, want %v (output %q)", c.in, warned, c.wantWarn, out)
			}
		})
	}
}

func TestIsSetWarnEmptyNamesTheProperty(t *testing.T) {
	out := captureStderr(t, func() {
		isSetWarnEmpty([]any{}, "sh -c 'echo boom >&2': command.stderr")
	})

	for _, want := range []string{
		"WARNING:",
		"sh -c 'echo boom >&2': command.stderr",
		"empty list",
		"always passes",
		`use "" to assert empty content`,
	} {
		if !strings.Contains(strings.ToLower(out), strings.ToLower(want)) {
			t.Errorf("warning %q is missing %q", out, want)
		}
	}
	if n := strings.Count(out, "\n"); n != 1 {
		t.Errorf("expected exactly one warning line, got %d in %q", n, out)
	}
}
