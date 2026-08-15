package system

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/goss-org/goss/util"
)

type trackingReadCloser struct {
	io.Reader
	closed bool
}

func (r *trackingReadCloser) Close() error {
	r.closed = true
	return nil
}

func TestDefHTTPCloseClosesBodyAfterBodyRead(t *testing.T) {
	body := &trackingReadCloser{Reader: strings.NewReader("ok")}
	httpResource := &DefHTTP{
		loaded: true,
		resp: &http.Response{
			Body: body,
		},
	}

	reader, err := httpResource.Body()
	if err != nil {
		t.Fatalf("Body() returned error: %v", err)
	}
	if _, err := io.ReadAll(reader); err != nil {
		t.Fatalf("ReadAll() returned error: %v", err)
	}
	if err := httpResource.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
	if !body.closed {
		t.Fatal("Close() did not close the response body")
	}
}

// TestNewDefHTTPRequestHeaderParsing pins that request-headers entries are
// parsed on the first colon, so the HTTP wire form "Name:value" works, and
// that an entry with no colon is skipped instead of panicking.
func TestNewDefHTTPRequestHeaderParsing(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"colon and space", "X-Foo: bar", "bar"},
		{"colon no space", "X-Foo:bar", "bar"},
		{"no colon", "X-Foo", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewDefHTTP(context.Background(), "http://example.com", nil, util.Config{
				RequestHeader: []string{tc.header},
			})
			def, ok := h.(*DefHTTP)
			if !ok {
				t.Fatalf("NewDefHTTP returned %T, want *DefHTTP", h)
			}
			if got := def.RequestHeader.Get("X-Foo"); got != tc.want {
				t.Errorf("RequestHeader.Get(%q) = %q, want %q", "X-Foo", got, tc.want)
			}
		})
	}
}
