package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type ContainSubstringMatcher struct {
	matchers.ContainSubstringMatcher
}

func ContainSubstring(substr string, args ...any) GossMatcher {
	return &ContainSubstringMatcher{
		matchers.ContainSubstringMatcher{
			Substr: substr,
			Args:   args,
		},
	}
}

func (m *ContainSubstringMatcher) FailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to contain substring",
		Expected: m.Substr,
	}
}

func (m *ContainSubstringMatcher) NegatedFailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to contain substring",
		Expected: m.Substr,
	}
}

func (m *ContainSubstringMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]any)
	j["contain-substring"] = m.Substr
	return json.Marshal(j)
}
