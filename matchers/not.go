package matchers

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

func (matcher *NotMatcher) FailureResult(actual interface{}) MatcherResult {
	return matcher.Matcher.NegatedFailureResult(actual)
}

func (matcher *NotMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return matcher.Matcher.FailureResult(actual)
}

// FIXME: wtf
func (matcher *NotMatcher) String() string {
	return Object(matcher, 0)
}
