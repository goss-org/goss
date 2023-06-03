package matchers

import (
	"encoding/json"
)

type AndMatcher struct {
	fakeOmegaMatcher
	Matchers []GossMatcher

	// state
	firstFailedMatcher GossMatcher
}

func And(ms ...GossMatcher) GossMatcher {
	return &AndMatcher{Matchers: ms}
}

func (m *AndMatcher) Match(actual interface{}) (success bool, err error) {
	m.firstFailedMatcher = nil
	for _, matcher := range m.Matchers {
		success, err := matcher.Match(actual)
		if !success || err != nil {
			m.firstFailedMatcher = matcher
			return false, err
		}
	}
	return true, nil
}

func (m *AndMatcher) FailureResult(actual interface{}) MatcherResult {
	return m.firstFailedMatcher.FailureResult(actual)
}

func (m *AndMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "To not satisfy all of these matchers",
		Expected: m.Matchers,
	}
}

func (m *AndMatcher) MarshalJSON() ([]byte, error) {
	if len(m.Matchers) == 1 {
		return json.Marshal(m.Matchers[0])
	}
	j := make(map[string]interface{})
	j["and"] = m.Matchers
	return json.Marshal(j)
}
