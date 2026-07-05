package system

import (
	"io"
	"net/http"
	"strings"
	"testing"
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
