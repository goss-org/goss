package outputs

import (
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

// logTrace emits a per-test trace line via logger. Outputers call this for
// every individual test result; accepting the logger as a parameter keeps
// logTrace free of hidden dependencies on any process-wide log sink.
func logTrace(logger util.Logger, level string, msg string, testResult resource.TestResult, withIntResult bool) {
	if withIntResult {
		logger.Printf("[%s] %s: %s => %s (%s %+v %+v) [%.02f] [%d]",
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
		logger.Printf("[%s] %s: %s => %s (%s %+v %+v) [%.02f]",
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
