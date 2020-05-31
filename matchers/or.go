package matchers

import "github.com/onsi/gomega/format"

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

func (matcher *OrMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "To satisfy at least one of these matchers",
		Expected: matcher.Matchers,
	}
}

func (matcher *OrMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	firstSuccessfulMatcher := getUnexported(matcher, "firstSuccessfulMatcher")
	return firstSuccessfulMatcher.(GossMatcher).NegatedFailureResult(actual)
}

func (matcher *OrMatcher) String() string {
	return format.Object(matcher, 0)
}
