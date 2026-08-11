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

// FailureResult reports the first matcher that failed. firstFailedMatcher is
// only set once Match has actually run a child matcher, and that does not
// always happen: WithSafeTransformMatcher.Match returns early when its
// transform errors — a gjson path that does not exist, for example — so the
// wrapped And never evaluates anything and its state stays nil. Dereferencing
// it there panicked (#982). Report the And itself in that case, which keeps the
// surrounding transform chain and raw value in the output so the user can see
// which path they actually matched against.
func (m *AndMatcher) FailureResult(actual any) MatcherResult {
	if m.firstFailedMatcher == nil {
		return MatcherResult{
			Actual:   actual,
			Message:  "to satisfy all of these matchers",
			Expected: m.Matchers,
		}
	}
	return m.firstFailedMatcher.FailureResult(actual)
}

func (m *AndMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to satisfy all of these matchers",
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
