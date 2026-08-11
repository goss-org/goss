package goss

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCacheMissLogIsGatedByLogLevel is the regression test for
// https://github.com/goss-org/goss/issues/991: the cache-miss message emitted
// on every health probe could not be muted at any --loglevel, because it was
// logged without a level prefix for logutils to match on.
func TestCacheMissLogIsGatedByLogLevel(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		level   string
		visible bool
	}{
		"default INFO mutes it": {level: "INFO", visible: false},
		"WARN mutes it":         {level: "WARN", visible: false},
		"ERROR mutes it":        {level: "ERROR", visible: false},
		"DEBUG shows it":        {level: "DEBUG", visible: true},
		"TRACE shows it":        {level: "TRACE", visible: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			filter, err := newLogFilter(tc.level, &buf)
			require.NoError(t, err)

			// Deliberately a local logger: the message must be gated by the
			// filter itself, not by anything the global logger happens to do.
			log.New(filter, "", 0).Printf(cacheMissLogFormat, "res")

			if tc.visible {
				require.Contains(t, buf.String(), "Stale cache[res], running tests",
					"message should be emitted at level %s", tc.level)
			} else {
				require.Empty(t, buf.String(),
					"message should be muted at level %s", tc.level)
			}
		})
	}
}

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
