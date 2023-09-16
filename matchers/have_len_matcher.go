package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type HaveLenMatcher struct {
	matchers.HaveLenMatcher
}

func HaveLen(count int) GossMatcher {
	return &HaveLenMatcher{
		matchers.HaveLenMatcher{
			Count: count,
		},
	}
}

func (m *HaveLenMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have length",
		Expected: m.Count,
	}
}

func (m *HaveLenMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have length",
		Expected: m.Count,
	}
}

func (m *HaveLenMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["have-len"] = m.Count
	return json.Marshal(j)
}
