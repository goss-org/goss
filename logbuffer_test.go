package goss

import (
	"bytes"
	"sync"
)

// syncBuffer is a bytes.Buffer that is safe to write and read concurrently.
//
// The serve tests capture log output by pointing the process-wide logger at a
// buffer with log.SetOutput. That destination is global, so a parallel test
// still writes into whichever buffer was installed last while its owner reads
// it — a data race on the buffer even though each test declares its own.
// Guarding the buffer removes the race without giving up the shared logger.
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

func (b *syncBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}
