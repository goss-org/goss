package goss

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The cache-miss regression test that lived here has moved to
// serve_logging_test.go, where it now asserts on structured records: the
// message it used to match on carried its level as a text prefix, and no longer
// exists.

// TestNewLogFilterRejectsUnknownLevel pins the error path, and documents that
// FATAL is not among the supported levels.
func TestNewLogFilterRejectsUnknownLevel(t *testing.T) {
	t.Parallel()

	for _, level := range []string{"FATAL", "", "verbose"} {
		_, err := newLogFilter(level, &bytes.Buffer{})
		require.Error(t, err, "level %q should be rejected", level)
	}
}

// TestNewLogFilterAcceptsSupportedLevels guards the level list documented in
// docs/cli.md against drift, and pins both the case-insensitive handling and
// the "lower levels include the upper ones" contract that docs/cli.md promises.
//
// This asserts on what actually reaches the writer rather than on the filter's
// own fields, so it keeps its meaning if the logging backend is ever swapped
// out for something that models levels differently.
func TestNewLogFilterAcceptsSupportedLevels(t *testing.T) {
	t.Parallel()

	// Ordered most to least verbose: configuring levels[i] must emit exactly
	// levels[i:] and mute everything before it.
	levels := []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

	for i, configured := range levels {
		for _, spelling := range []string{configured, strings.ToLower(configured)} {
			t.Run(spelling, func(t *testing.T) {
				t.Parallel()

				var buf bytes.Buffer
				filter, err := newLogFilter(spelling, &buf)
				require.NoError(t, err, "level %q should be accepted", spelling)

				logger := log.New(filter, "", 0)
				for _, emitted := range levels {
					logger.Printf("[%s] %s-message", emitted, emitted)
				}

				got := buf.String()
				for j, emitted := range levels {
					if j >= i {
						require.Contains(t, got, emitted+"-message",
							"%s should be emitted when configured at %s", emitted, configured)
					} else {
						require.NotContains(t, got, emitted+"-message",
							"%s should be muted when configured at %s", emitted, configured)
					}
				}
			})
		}
	}
}
