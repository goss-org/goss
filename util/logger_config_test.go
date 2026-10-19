package util

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWithLogger checks that the logger is injectable through the same option
// pattern as the rest of the runtime configuration.
func TestWithLogger(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))

	c, err := NewConfig(WithLogger(logger))
	require.NoError(t, err)
	require.Same(t, logger, c.Logger)

	LoggerOrDiscard(c.Logger).Info("marker")
	require.Contains(t, buf.String(), "marker")
}

// TestNewConfigDoesNotDefaultTheLogger pins the central decision behind this
// change: with no logger supplied the library has nowhere to write, and that is
// deliberate. Defaulting to slog.Default() here would put records somewhere the
// embedder did not choose, which is the defect being fixed.
func TestNewConfigDoesNotDefaultTheLogger(t *testing.T) {
	t.Parallel()

	c, err := NewConfig()
	require.NoError(t, err)
	require.Nil(t, c.Logger)

	require.NotPanics(t, func() {
		LoggerOrDiscard(c.Logger).Error("discarded")
	})
}

// TestWithLoggerAcceptsNil documents that passing nil is not an error: it leaves
// the config in exactly the state the documented composite-literal pattern
// produces, which every derivation already has to survive.
func TestWithLoggerAcceptsNil(t *testing.T) {
	t.Parallel()

	c, err := NewConfig(WithLogger(nil))
	require.NoError(t, err)
	require.Nil(t, c.Logger)
}

// TestOutputConfigLogger covers the outputs/ injection point: an outputer reads
// its logger from the OutputConfig, and a zero value stays safe because
// embedders build that struct as a literal too.
func TestOutputConfigLogger(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))

	oc := OutputConfig{Logger: logger}
	require.Same(t, logger, oc.Logger)

	require.Nil(t, OutputConfig{}.Logger)
	require.NotPanics(t, func() {
		LoggerOrDiscard(OutputConfig{}.Logger).Error("discarded")
	})
}
