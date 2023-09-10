package goss

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/goss-org/goss/util"
	"github.com/hashicorp/logutils"
)

func setLogLevel(c *util.Config) error {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
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
