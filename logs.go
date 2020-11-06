package goss

import (
	"fmt"
	"log"
	"os"

	"github.com/goss-org/goss/util"
	"github.com/hashicorp/logutils"
)

func setLogLevel(c *util.Config) error {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	logLevelFound := false
	for _, lvl := range filter.Levels {
		if string(lvl) == c.LogLevel {
			logLevelFound = true
			break
		}
	}
	if !logLevelFound {
		return fmt.Errorf("Unsupported log level: %s", c.LogLevel)
	}
	filter.MinLevel = logutils.LogLevel(c.LogLevel)
	log.Printf("Setting log level to %v", c.LogLevel)
	log.SetOutput(filter)
	return nil
}
