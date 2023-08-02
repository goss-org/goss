package matchers

import (
	"encoding/json"
)

type OrMatcher struct {
	fakeOmegaMatcher

	Matchers []GossMatcher

	// state
	firstSuccessfulMatcher GossMatcher
}

func Or(ms ...GossMatcher) GossMatcher {
	return &OrMatcher{Matchers: ms}
}

func (m *OrMatcher) Match(actual interface{}) (success bool, err error) {
	m.firstSuccessfulMatcher = nil
	for _, matcher := range m.Matchers {
		success, err := matcher.Match(actual)
		if err != nil {
			return false, err
		}
		if success {
			m.firstSuccessfulMatcher = matcher
			return true, nil
		}
	}
	return false, nil
}

func (m *OrMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to satisfy at least one of these matchers",
		Expected: m.Matchers,
	}
}

func (m *OrMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	firstSuccessfulMatcher := getUnexported(m, "firstSuccessfulMatcher")
	return firstSuccessfulMatcher.(GossMatcher).NegatedFailureResult(actual)
}

func (m *OrMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["or"] = m.Matchers
	return json.Marshal(j)
}
