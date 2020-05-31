package matchers

import (
	"github.com/onsi/gomega/matchers"
)

type MatchRegexpMatcher struct {
	matchers.MatchRegexpMatcher
}

func MatchRegexp(regexp string, args ...interface{}) GossMatcher {
	return &MatchRegexpMatcher{
		matchers.MatchRegexpMatcher{
			Regexp: regexp,
			Args:   args,
		},
	}
}

func (matcher *MatchRegexpMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to match regular expression",
		Expected: matcher.Regexp,
	}
}

func (matcher *MatchRegexpMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to match regular expression",
		Expected: matcher.Regexp,
	}
}

func (matcher *MatchRegexpMatcher) String() string {
	return Object(matcher.MatchRegexpMatcher, 0)
}
