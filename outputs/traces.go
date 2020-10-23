package outputs

import (
	"log"

	"github.com/aelsabbahy/goss/resource"
)

func logTrace(level string, msg string, testResult resource.TestResult) {
	log.Printf("[%s] %s: %s => %s (%s %+v %+v) [%.02f]",
		level,
		msg,
		testResult.ResourceType,
		testResult.ResourceId,
		testResult.Property,
		testResult.Expected,
		testResult.Found,
		testResult.Duration.Seconds(),
	)
}
