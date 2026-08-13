package util

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"testing"
)

// Logger is the minimal logging seam used throughout goss. It abstracts away
// the standard library log package to enable:
//
//  1. Parallel test execution with per-test log capture (the standard library
//     log package exposes a single process-wide default logger whose output
//     writer can only be swapped atomically, which races under t.Parallel).
//  2. Custom log sinks (files, structured writers, etc.) without requiring
//     callers to mutate global state.
//  3. Test assertions on logged output without calling log.SetOutput.
//
// The interface intentionally mirrors the two standard library log functions
// most used by goss (Printf, Fatalf) so that migrating call sites requires
// only prefixing them with an accessor (e.g. "c.Log().Printf(...)").
// Log levels continue to be expressed as message prefixes ("[DEBUG] ...",
// "[TRACE] ...") and filtered by hashicorp/logutils, matching the
// pre-refactor convention.
type Logger interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

// DefaultLogger is the production Logger. It delegates to the standard
// library log package, preserving byte-identical output with the pre-refactor
// implementation. It is the zero-value fallback returned by Config.Log when
// no logger has been explicitly injected.
type DefaultLogger struct{}

// Printf delegates to log.Printf.
func (DefaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Fatalf delegates to log.Fatalf (logs the message then calls os.Exit(1)).
func (DefaultLogger) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}

// TestLogger is a goroutine-safe Logger for tests. It forwards messages to
// testing.T.Log (which the Go runtime documents as safe for concurrent use
// from any goroutine associated with the test) and also accumulates them in
// an internal buffer so tests can assert on log content with helpers like
// assert.Contains(t, tl.String(), "expected substring").
//
// Construct via NewTestLogger. The zero value is not usable.
type TestLogger struct {
	t   *testing.T
	mu  sync.Mutex
	buf bytes.Buffer
}

// NewTestLogger returns a TestLogger bound to the given *testing.T.
func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t: t}
}

// Printf formats according to a format specifier and records the result.
// The message is appended to the internal buffer (followed by a newline) and
// also emitted via t.Log so it appears in "-v" output next to the test that
// produced it. Safe for concurrent use.
func (tl *TestLogger) Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	tl.mu.Lock()
	tl.buf.WriteString(msg)
	tl.buf.WriteByte('\n')
	tl.mu.Unlock()
	// t.Log is documented safe for concurrent use by goroutines associated
	// with the test. It does not panic after the test has completed; it
	// simply writes to the test's buffered log.
	tl.t.Log(msg)
}

// Fatalf formats the message and fails the current test via t.Fatalf. It
// does NOT call os.Exit (unlike log.Fatalf), which would tear down the whole
// test binary. Safe for concurrent use.
func (tl *TestLogger) Fatalf(format string, v ...interface{}) {
	// t.Fatalf is documented safe for concurrent use from the test goroutine
	// and goroutines created during the test.
	tl.t.Fatalf(format, v...)
}

// String returns the accumulated log output. Safe for concurrent use.
func (tl *TestLogger) String() string {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	return tl.buf.String()
}

// Reset discards any accumulated log output. Useful between phases of a test
// that wants to assert on freshly-produced messages (e.g. separating
// cache-miss from cache-hit assertions).
func (tl *TestLogger) Reset() {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.buf.Reset()
}
