package outputs

import (
	"context"
	"log/slog"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

// Outcomes as they appear in the outcome attribute of a validation result.
const (
	outcomeSuccess = "success"
	outcomeFail    = "fail"
	outcomeSkip    = "skip"
)

// Statuses as they appear in the status attribute of a validation summary.
const (
	statusOK   = "ok"
	statusFail = "fail"
)

// outcomeFor maps a result to the outcome attribute. Skipped is checked as well
// as the result code because the two are set independently: a skipped result
// carries resource.SKIP, but a result can also be flagged skipped on its own.
func outcomeFor(testResult resource.TestResult) string {
	switch {
	case testResult.Result == resource.FAIL:
		return outcomeFail
	case testResult.Result == resource.SKIP || testResult.Skipped:
		return outcomeSkip
	default:
		return outcomeSuccess
	}
}

// logTrace emits one validation result at TRACE.
//
// withIntResult preserves the difference between the two callers: the json
// outputer's records carry the numeric result and rspecish's do not.
func logTrace(logger *slog.Logger, outcome string, testResult resource.TestResult, withIntResult bool) {
	l := util.LoggerOrDiscard(logger)
	if !l.Enabled(context.Background(), util.LevelTrace) {
		return
	}
	attrs := [14]any{
		"outcome", outcome,
		"resource_type", testResult.ResourceType,
		"resource_id", testResult.ResourceId,
		"property", testResult.Property,
		// Expected and actual stay native values rather than being formatted
		// here, so that a handler decides how to render them.
		slog.Any("expected", testResult.MatcherResult.Expected),
		slog.Any("actual", testResult.MatcherResult.Actual),
		"duration_seconds", testResult.Duration.Seconds(),
	}

	if withIntResult {
		attrs[12], attrs[13] = "result", int(testResult.Result)
		util.Trace(l, "validation result", attrs[:14]...)
	} else {
		util.Trace(l, "validation result", attrs[:12]...)
	}
}
