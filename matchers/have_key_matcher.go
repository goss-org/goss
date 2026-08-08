package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type HaveKeyMatcher struct {
	matchers.HaveKeyMatcher
}

func HaveKey(key any) GossMatcher {
	return &HaveKeyMatcher{
		matchers.HaveKeyMatcher{
			Key: key,
		},
	}
}

func (m *HaveKeyMatcher) FailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have key matching",
		Expected: m.Key,
	}
}

func (m *HaveKeyMatcher) NegatedFailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have key matching",
		Expected: m.Key,
	}
}

func (m *HaveKeyMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]any)
	j["have-key"] = m.Key
	return json.Marshal(j)
}
