package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type HaveKeyMatcher struct {
	matchers.HaveKeyMatcher
}

func HaveKey(key interface{}) GossMatcher {
	return &HaveKeyMatcher{
		matchers.HaveKeyMatcher{
			Key: key,
		},
	}
}

func (m *HaveKeyMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have key matching",
		Expected: m.Key,
	}
}

func (m *HaveKeyMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have key matching",
		Expected: m.Key,
	}
}

func (m *HaveKeyMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["have-key"] = m.Key
	return json.Marshal(j)
}

func (m *HaveKeyMatcher) String() string {
	return Object(m.HaveKeyMatcher, 0)
}
