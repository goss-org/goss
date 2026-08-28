package goss

import (
	"bytes"
	"sync"
)

// syncBuffer is a bytes.Buffer that is safe to write and read concurrently.
//
// Each test now owns the logger it captures, so buffers are no longer shared
// between tests. The locking is still needed within one test: goss logs from
// the goroutines serving requests, and slog only serialises writes from its own
// side, so a test reading the buffer while a request is in flight races it.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
