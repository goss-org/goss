package util

import (
	"sync"

	"github.com/fatih/color"
)

// colorOnce ensures that color.NoColor is set exactly once, avoiding a data
// race when multiple goroutines (for example, concurrent HTTP requests served
// by goss in server mode, or parallel output writers) try to mutate the
// package-level color.NoColor variable from github.com/fatih/color.
//
// color.NoColor is a boolean that controls whether color output is disabled.
// Historically, goss code set it directly from several entry points which
// meant concurrent callers could race on the write. For goss's purposes the
// value only needs to be decided once, at process startup, based on the
// user's configuration (CLI flag, config, or output format).
//
// This lives in the util package (rather than goss or outputs) so both
// packages can share a single sync.Once instance -- otherwise two
// independent sync.Once guards would each perform a write to color.NoColor
// and racy reads from color formatting could observe one-or-the-other.
var colorOnce sync.Once

// InitNoColor sets color.NoColor to the given value exactly once per process.
// Subsequent calls are no-ops. This is safe to call from multiple goroutines.
//
// Callers that want to explicitly disable color (e.g. machine-readable output
// formats such as JSON or JUnit) should call InitNoColor(true). Callers that
// want to respect the terminal's default should not call this function.
func InitNoColor(disable bool) {
	colorOnce.Do(func() {
		color.NoColor = disable
	})
}
