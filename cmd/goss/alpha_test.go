package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAlphaRejected covers the platform half of the alpha gate. The predicate
// is a pure function of the OS name and the flag value so that the darwin and
// windows decisions are testable from a Linux CI machine, and the Linux
// decision is testable from a developer's Mac.
func TestAlphaRejected(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		goos     string
		useAlpha string
		want     bool
	}{
		"darwin without the opt-in":    {goos: "darwin", useAlpha: "0", want: true},
		"darwin with an empty opt-in":  {goos: "darwin", useAlpha: "", want: true},
		"darwin with a wrong opt-in":   {goos: "darwin", useAlpha: "true", want: true},
		"darwin with the opt-in":       {goos: "darwin", useAlpha: "1", want: false},
		"windows without the opt-in":   {goos: "windows", useAlpha: "0", want: true},
		"windows with the opt-in":      {goos: "windows", useAlpha: "1", want: false},
		"linux is never rejected":      {goos: "linux", useAlpha: "", want: false},
		"linux ignores the opt-in":     {goos: "linux", useAlpha: "0", want: false},
		"freebsd is never rejected":    {goos: "freebsd", useAlpha: "0", want: false},
		"an unknown OS is not blocked": {goos: "plan9", useAlpha: "", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, alphaRejected(tc.goos, tc.useAlpha))
		})
	}
}

// TestAlphaMessage covers the diagnostic half of the alpha gate, pinning the
// complete text including the per-platform guidance. This is the message the
// terminal path prints, so its exact bytes are part of the CLI's behaviour.
func TestAlphaMessage(t *testing.T) {
	t.Parallel()

	t.Run("darwin", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, `Terminating.

To bypass this and use the binary anyway:

export GOSS_USE_ALPHA=1`, alphaMessage("darwin"))
	})

	t.Run("windows", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, `Terminating.

To bypass this and use the binary anyway:

In cmd:        set GOSS_USE_ALPHA=1
In powershell: $env:GOSS_USE_ALPHA=1
In bash:       export GOSS_USE_ALPHA=1`, alphaMessage("windows"))
	})

	// The gate never reaches this on any other platform, so the message has no
	// guidance to give. Asserting it documents that the map lookup miss is
	// harmless rather than leaving it as an accident waiting to be discovered.
	t.Run("any other OS has no guidance", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, `Terminating.

To bypass this and use the binary anyway:

`, alphaMessage("linux"))
	})
}
