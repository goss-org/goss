package outputs

import (
	"log"

	"github.com/goss-org/goss/resource"
)

func logTrace(level string, msg string, testResult resource.TestResult, withIntResult bool) {
	if withIntResult {
		log.Printf("[%s] %s: %s => %s (%s %+v %+v) [%.02f] [%d]",
			level,
			msg,
			testResult.ResourceType,
			testResult.ResourceId,
			testResult.Property,
			testResult.Expected,
			testResult.Found,
			testResult.Duration.Seconds(),
			testResult.Result,
		)
	} else {
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
}
