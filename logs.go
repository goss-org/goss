package goss

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/hashicorp/logutils"
)

func setLogLevel(c *util.Config) error {
	filter, err := newLogFilter(c.LogLevel, os.Stderr)
	if err != nil {
		return err
	}
	log.SetFlags(0) // Turn off standard timestamp flags
	log.SetOutput(&timestampedWriter{filter})
	log.Printf("[DEBUG] Setting log level to %v", strings.ToUpper(c.LogLevel))
	return nil
}

// newLogFilter builds the level filter used to gate goss's log output. Levels
// are expressed as a prefix on each message (for example "[DEBUG] ..."), so any
// message logged without one is emitted regardless of the configured level.
func newLogFilter(level string, w io.Writer) (*logutils.LevelFilter, error) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   w,
	}
	want := strings.ToUpper(level)
	for _, lvl := range filter.Levels {
		if string(lvl) == want {
			filter.MinLevel = lvl
			return filter, nil
		}
	}
	return nil, fmt.Errorf("Unsupported log level: %s", level)
}

type timestampedWriter struct {
	wrappedWriter io.Writer
}

func (t *timestampedWriter) Write(b []byte) (int, error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	return fmt.Fprintf(t.wrappedWriter, "%s %s", timestamp, b)
}
