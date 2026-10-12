package matchers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Not sends FailureResult to the Or's NegatedFailureResult.
func TestOrMatcherNegatedFailureResultWhenNoChildMatcherRan(t *testing.T) {
	inner := Or(HaveKey("a"), HaveKey("b"))
	matcher := WithSafeTransform(erroringTransform{}, Not(inner))

	success, err := matcher.Match(`{"this": {"is": {"just": {"a": "test"}}}}`)
	assert.False(t, success)
	assert.Error(t, err)

	assert.NotPanics(t, func() {
		result := matcher.FailureResult("actual")
		assert.Equal(t, "not to satisfy any of these matchers", result.Message)
	})
}

func TestOrMatcherNegatedFailureResultWhenAChildErrors(t *testing.T) {
	matcher := Not(Or(HaveKey("a"), HaveKey("b")))

	success, err := matcher.Match("not a map")
	assert.False(t, success)
	assert.Error(t, err)

	assert.NotPanics(t, func() {
		result := matcher.FailureResult("not a map")
		assert.Equal(t, "not to satisfy any of these matchers", result.Message)
	})
}

func TestOrMatcherNegatedFailureResultStillReportsTheMatchedChild(t *testing.T) {
	m := Or(HaveKey("missing"), HaveKey("present"))

	success, err := m.Match(map[string]any{"present": 1})
	assert.True(t, success)
	assert.NoError(t, err)

	result := m.NegatedFailureResult(map[string]any{"present": 1})
	assert.Equal(t, "not to have key matching", result.Message)
	assert.Equal(t, "present", result.Expected)
}
