package util

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoggerInterfaceSatisfied is a compile-time check. If either
// DefaultLogger or *TestLogger stops satisfying Logger, this file will fail
// to compile.
var (
	_ Logger = DefaultLogger{}
	_ Logger = (*TestLogger)(nil)
)

func TestDefaultLogger_Printf(t *testing.T) {
	// Capture the standard library log output via log.SetOutput, restoring
	// the original sink afterwards so this test does not bleed into others.
	var buf bytes.Buffer
	origFlags := log.Flags()
	origOutput := log.Writer()
	log.SetFlags(0) // suppress timestamps for stable matching
	log.SetOutput(&buf)
	defer func() {
		log.SetFlags(origFlags)
		log.SetOutput(origOutput)
	}()

	d := DefaultLogger{}
	d.Printf("hello %s %d", "world", 42)

	assert.Equal(t, "hello world 42\n", buf.String())
}

func TestNewTestLogger_Printf_Records(t *testing.T) {
	t.Parallel()
	tl := NewTestLogger(t)
	tl.Printf("hello %s", "world")
	tl.Printf("second line: %d", 7)

	got := tl.String()
	assert.Contains(t, got, "hello world")
	assert.Contains(t, got, "second line: 7")
	// Each Printf appends a newline.
	assert.Equal(t, 2, strings.Count(got, "\n"))
}

func TestTestLogger_Reset(t *testing.T) {
	t.Parallel()
	tl := NewTestLogger(t)
	tl.Printf("before reset")
	assert.Contains(t, tl.String(), "before reset")

	tl.Reset()
	assert.Equal(t, "", tl.String())

	tl.Printf("after reset")
	got := tl.String()
	assert.NotContains(t, got, "before reset")
	assert.Contains(t, got, "after reset")
}

// TestTestLogger_ConcurrentWrites exercises the mutex guarding the buffer.
// Run under -race, this asserts no data race and no lost writes.
func TestTestLogger_ConcurrentWrites(t *testing.T) {
	t.Parallel()
	tl := NewTestLogger(t)

	const goroutines = 50
	const perGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				tl.Printf("g=%d j=%d", i, j)
			}
		}()
	}
	wg.Wait()

	// Every message must appear exactly once; no writes lost to races.
	// Compare against the set of lines rather than substring-search, since
	// "g=4 j=1" is a substring of "g=48 j=1".
	lines := strings.Split(strings.TrimRight(tl.String(), "\n"), "\n")
	seen := make(map[string]int, goroutines*perGoroutine)
	for _, line := range lines {
		seen[line]++
	}
	assert.Equal(t, goroutines*perGoroutine, len(lines), "expected %d total lines", goroutines*perGoroutine)
	for i := 0; i < goroutines; i++ {
		for j := 0; j < perGoroutine; j++ {
			want := fmt.Sprintf("g=%d j=%d", i, j)
			assert.Equal(t, 1, seen[want], "expected exactly one occurrence of %q", want)
		}
	}
}

func TestConfig_Log_NilFallback(t *testing.T) {
	t.Parallel()
	cfg := &Config{}

	got := cfg.Log()
	require.NotNil(t, got, "Config.Log() must never return nil")

	// Must be usable without panicking.
	assert.NotPanics(t, func() {
		got.Printf("smoke test %d", 1)
	})
}

func TestConfig_Log_CustomLogger(t *testing.T) {
	t.Parallel()
	tl := NewTestLogger(t)
	cfg := &Config{Logger: tl}

	assert.Same(t, tl, cfg.Log(), "Config.Log() must return the injected logger verbatim")
}

func TestWithLogger_SetsField(t *testing.T) {
	t.Parallel()
	tl := NewTestLogger(t)

	cfg, err := NewConfig(WithLogger(tl))
	require.NoError(t, err)
	assert.Same(t, tl, cfg.Logger)
	assert.Same(t, tl, cfg.Log())
}

// TestDefaultLogger_UsedWhenConfigLoggerNil ensures that calling Config.Log()
// on a freshly-built Config (no WithLogger applied) yields a DefaultLogger,
// not a zero value or nil.
func TestDefaultLogger_UsedWhenConfigLoggerNil(t *testing.T) {
	t.Parallel()
	cfg, err := NewConfig()
	require.NoError(t, err)
	require.Nil(t, cfg.Logger, "precondition: Config.Logger should be unset")

	_, ok := cfg.Log().(DefaultLogger)
	assert.True(t, ok, "expected DefaultLogger, got %T", cfg.Log())
}
