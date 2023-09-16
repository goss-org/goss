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
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetFlags(0) // Turn off standard timestamp flags
	log.SetOutput(&timestampedWriter{filter})
	for _, lvl := range filter.Levels {
		cLvl := strings.ToUpper(c.LogLevel)
		if string(lvl) == cLvl {
			filter.MinLevel = lvl
			log.Printf("[DEBUG] Setting log level to %v", cLvl)
			return nil
		}
	}
	return fmt.Errorf("Unsupported log level: %s", c.LogLevel)
}

type timestampedWriter struct {
	wrappedWriter io.Writer
}

func (t *timestampedWriter) Write(b []byte) (int, error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	return fmt.Fprintf(t.wrappedWriter, "%s %s", timestamp, b)
}
