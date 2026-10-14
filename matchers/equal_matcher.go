package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type EqualMatcher struct {
	matchers.EqualMatcher
}

func Equal(element any) GossMatcher {
	return &EqualMatcher{
		matchers.EqualMatcher{
			Expected: element,
		},
	}
}

func (m *EqualMatcher) FailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to equal",
		Expected: m.Expected,
	}
}

func (m *EqualMatcher) NegatedFailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to equal",
		Expected: m.Expected,
	}
}

func (m *EqualMatcher) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Expected)
}
