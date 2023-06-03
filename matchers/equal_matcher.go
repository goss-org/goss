package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type EqualMatcher struct {
	matchers.EqualMatcher
}

func Equal(element interface{}) GossMatcher {
	return &EqualMatcher{
		matchers.EqualMatcher{
			Expected: element,
		},
	}
}

func (m *EqualMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to equal",
		Expected: m.Expected,
	}
}

func (m *EqualMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to equal",
		Expected: m.Expected,
	}
}

func (m *EqualMatcher) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Expected)
	j := make(map[string]interface{})
	j["equal"] = m.Expected
	return json.Marshal(j)
}
