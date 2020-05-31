package matchers

import (
	"github.com/onsi/gomega/matchers"
)

type HaveKeyMatcher struct {
	matchers.HaveKeyMatcher
}

func HaveKey(key interface{}) GossMatcher {
	return &HaveKeyMatcher{
		matchers.HaveKeyMatcher{
			Key: key,
		},
	}
}

func (matcher *HaveKeyMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have key matching",
		Expected: matcher.Key,
	}
}

func (matcher *HaveKeyMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have key matching",
		Expected: matcher.Key,
	}
}

func (matcher *HaveKeyMatcher) String() string {
	return Object(matcher.HaveKeyMatcher, 0)
}
