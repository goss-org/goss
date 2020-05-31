package matchers

import (
	"github.com/onsi/gomega/matchers"
)

type ContainElementMatcher struct {
	matchers.ContainElementMatcher
}

func ContainElement(element interface{}) GossMatcher {
	return &ContainElementMatcher{
		matchers.ContainElementMatcher{
			Element: element,
		},
	}
}

func (matcher *ContainElementMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to contain element matching",
		Expected: matcher.Element,
	}
}

func (matcher *ContainElementMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to contain element matching",
		Expected: matcher.Element,
	}
}

func (matcher *ContainElementMatcher) String() string {
	return Object(matcher.ContainElementMatcher, 0)
}
