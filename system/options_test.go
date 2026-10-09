package system

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewAcceptsALogger covers the structural half of the plumbing: the logger
// reaches the System, and the resources it creates read it back. Routing actual
// subject command output through it lands with the system/ conversion.
func TestNewAcceptsALogger(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	sys := New("", WithLogger(logger))

	require.Same(t, logger, sys.loggerOrDiscard())

	sys.loggerOrDiscard().Debug("marker")
	require.Contains(t, buf.String(), "marker")
}

// TestSystemLoggerIsNilSafe pins the nil-logger guard. System is constructible
// as a literal, New is callable without options, and neither may hand out a nil
// *slog.Logger, which panics on every method that matters.
func TestSystemLoggerIsNilSafe(t *testing.T) {
	t.Parallel()

	tests := map[string]*System{
		"nil System":          nil,
		"zero value literal":  {},
		"New with no option":  New(""),
		"New with nil logger": New("", WithLogger(nil)),
	}

	for name, sys := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				logger := sys.loggerOrDiscard()
				require.NotNil(t, logger)
				logger.With("key", "value").Debug("discarded")
			})
		})
	}
}

// TestNewKeepsPositionalCallsCompiling is the reason system.New takes variadic
// options rather than a second positional parameter: every existing caller
// keeps compiling, which keeps this a compatible change for embedders.
func TestNewKeepsPositionalCallsCompiling(t *testing.T) {
	t.Parallel()

	sys := New("")

	require.NotNil(t, sys.NewCommand)
	require.NotNil(t, sys.NewPackage)
	require.NotNil(t, sys.NewService)
}
