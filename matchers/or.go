package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/format"
)

type OrMatcher struct {
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
		Message:  "To satisfy at least one of these matchers",
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

func (m *OrMatcher) String() string {
	return format.Object(m, 0)
}

// FailureMessage is a stub to honor omegaMatcher interface
func (m *OrMatcher) FailureMessage(_ interface{}) (message string) {
	return ""
}

// NegatedFailureMessage is a stub to honor omegaMatcher interface
func (m *OrMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return ""
}
