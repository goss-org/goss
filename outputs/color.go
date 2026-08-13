package outputs

import (
	"sync"

	"github.com/fatih/color"
)

// color.NoColor is a package-level variable in github.com/fatih/color that is
// read every time one of the colorizing helpers below formats a string.
//
// Under `goss serve` the output formatters run concurrently, once per in-flight
// request, and the machine-readable formatters assign to color.NoColor on every
// invocation. That combination is a data race: unsynchronised writes from one
// request against reads from another.
//
// Guarding the assignment and the reads with a single RWMutex removes the race
// while preserving the existing last-writer-wins behaviour.
var colorMu sync.RWMutex

var (
	greenFn  = color.New(color.FgGreen).SprintfFunc()
	redFn    = color.New(color.FgRed).SprintfFunc()
	yellowFn = color.New(color.FgYellow).SprintfFunc()
)

// SetNoColor enables or disables ANSI colorization for all goss output. It is
// safe to call concurrently, and callers outside this package should use it in
// preference to assigning to color.NoColor directly.
func SetNoColor(disable bool) {
	colorMu.Lock()
	defer colorMu.Unlock()
	color.NoColor = disable
}

func green(format string, a ...any) string {
	return colorize(greenFn, format, a...)
}

func red(format string, a ...any) string {
	return colorize(redFn, format, a...)
}

func yellow(format string, a ...any) string {
	return colorize(yellowFn, format, a...)
}

func colorize(f func(string, ...any) string, format string, a ...any) string {
	colorMu.RLock()
	defer colorMu.RUnlock()
	return f(format, a...)
}
