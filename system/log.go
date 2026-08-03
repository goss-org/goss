package system

import (
	"bytes"

	"github.com/goss-org/goss/util"
)

// logBytes emits each non-empty line of b to logger with the given prefix.
// Returning no values (rather than accumulating a string) keeps the caller
// simple; the logger is injected at the call site rather than looked up via
// a package-level global, which keeps this function free of hidden side
// effects on any shared sink.
func logBytes(logger util.Logger, b []byte, prefix string) {
	if len(b) == 0 {
		return
	}
	lines := bytes.Split(b, []byte("\n"))
	for _, l := range lines {
		logger.Printf("[DEBUG]%s %s", prefix, l)
	}
}
