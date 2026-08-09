package matchers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// erroringTransform fails the way Gjson does when a path is absent.
type erroringTransform struct{}

func (erroringTransform) Transform(in any) (any, error) {
	return in, errors.New("Path not found: this.is.typo")
}

func (erroringTransform) TransformerType() string { return "erroringTransform" }

// A transform that errors makes WithSafeTransformMatcher.Match return before it
// ever calls the wrapped matcher, so the And never runs a child and its
// firstFailedMatcher stays nil. Asking it for a FailureResult used to panic
// (#982).
func TestAndMatcherFailureResultWhenNoChildMatcherRan(t *testing.T) {
	inner := And(HaveKey("a"))
	matcher := WithSafeTransform(erroringTransform{}, inner)

	success, err := matcher.Match(`{"this": {"is": {"just": {"a": "test"}}}}`)
	assert.False(t, success)
	assert.Error(t, err)

	assert.NotPanics(t, func() {
		result := matcher.FailureResult("actual")
		assert.Equal(t, "to satisfy all of these matchers", result.Message)
	})
}

// The nil result must not be returned once a child has genuinely failed —
// otherwise the guard would swallow the specific matcher that failed.
func TestAndMatcherFailureResultStillReportsTheFailedChild(t *testing.T) {
	m := And(HaveKey("missing"))

	success, err := m.Match(map[string]any{"present": 1})
	assert.False(t, success)
	assert.NoError(t, err)

	result := m.FailureResult(map[string]any{"present": 1})
	assert.Equal(t, "to have key matching", result.Message,
		"must delegate to the child that failed, not the And's own message")
}
