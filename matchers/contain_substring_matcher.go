package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type ContainSubstringMatcher struct {
	matchers.ContainSubstringMatcher
}

func ContainSubstring(substr string, args ...interface{}) GossMatcher {
	return &ContainSubstringMatcher{
		matchers.ContainSubstringMatcher{
			Substr: substr,
			Args:   args,
		},
	}
}

func (m *ContainSubstringMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to contain substring",
		Expected: m.Substr,
	}
}

func (m *ContainSubstringMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to contain substring",
		Expected: m.Substr,
	}
}

func (m *ContainSubstringMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["contain-substring"] = m.Substr
	return json.Marshal(j)
}
