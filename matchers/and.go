package matchers

import (
	"fmt"
)

type AndMatcher struct {
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

func (matcher *AndMatcher) FailureResult(actual interface{}) MatcherResult {
	return matcher.firstFailedMatcher.FailureResult(actual)
}

func (matcher *AndMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:  actual,
		Message: fmt.Sprintf("To not satisfy all of these matchers: %s", matcher.Matchers),
	}
}

//FIXME: Indentation is wrong
func (matcher *AndMatcher) String() string {
	return fmt.Sprintf("AndMatcher{Matchers:%v}", matcher.Matchers)
}
