package matchers

import (
	"github.com/onsi/gomega/matchers"
)

type HaveKeyWithValueMatcher struct {
	matchers.HaveKeyWithValueMatcher
}

func HaveKeyWithValue(key interface{}, value interface{}) GossMatcher {
	return &HaveKeyWithValueMatcher{
		matchers.HaveKeyWithValueMatcher{
			Key:   key,
			Value: value,
		},
	}
}

func (matcher *HaveKeyWithValueMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have {key: value} matching",
		Expected: matcher.Key,
	}
}

func (matcher *HaveKeyWithValueMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have {key: value} matching",
		Expected: matcher.Key,
	}
}

func (matcher *HaveKeyWithValueMatcher) String() string {
	return Object(matcher.HaveKeyWithValueMatcher, 0)
}
