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
			testResult.MatcherResult.Expected,
			testResult.MatcherResult.Actual,
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
			testResult.MatcherResult.Expected,
			testResult.MatcherResult.Actual,
			testResult.Duration.Seconds(),
		)
	}
}
