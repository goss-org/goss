package matchers

import (
	"encoding/json"
)

type NotMatcher struct {
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

// Stubs to match omegaMatcher
func (m *NotMatcher) FailureMessage(_ interface{}) (message string) {
	return ""
}

// Stubs to match omegaMatcher
func (m *NotMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return ""
}

func (m *NotMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["not"] = m.Matcher
	return json.Marshal(j)
}

// FIXME: wtf
func (m *NotMatcher) String() string {
	return Object(m, 0)
}
