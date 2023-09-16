package matchers

import (
	"encoding/json"
)

type NotMatcher struct {
	fakeOmegaMatcher
	Matcher GossMatcher
}

func Not(matcher GossMatcher) GossMatcher {
	return &NotMatcher{Matcher: matcher}
}

func (m *NotMatcher) Match(actual interface{}) (bool, error) {
	success, err := m.Matcher.Match(actual)
	if err != nil {
		return false, err
	}
	return !success, nil
}

func (m *NotMatcher) FailureResult(actual interface{}) MatcherResult {
	return m.Matcher.NegatedFailureResult(actual)
}

func (m *NotMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return m.Matcher.FailureResult(actual)
}

func (m *NotMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["not"] = m.Matcher
	return json.Marshal(j)
}
