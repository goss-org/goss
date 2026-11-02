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

func (m *OrMatcher) NegatedFailureResult(actual any) MatcherResult {
	// Unset when no child matched, e.g. a transform or a child errored (#1128).
	if m.firstSuccessfulMatcher == nil {
		return MatcherResult{
			Actual:   actual,
			Message:  "not to satisfy any of these matchers",
			Expected: m.Matchers,
		}
	}
	return m.firstSuccessfulMatcher.NegatedFailureResult(actual)
}

func (m *OrMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["or"] = m.Matchers
	return json.Marshal(j)
}
