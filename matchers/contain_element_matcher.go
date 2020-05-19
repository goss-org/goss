package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type ContainElementMatcher struct {
	matchers.ContainElementMatcher
}

func ContainElement(element interface{}) types.GomegaMatcher {
	return &ContainElementMatcher{
		matchers.ContainElementMatcher{
			Element: element,
		},
	}
}

func (matcher *ContainElementMatcher) FailureMessage(actual interface{}) (message string) {
	return Message(actual, "to contain element matching", matcher.Element)
}

func (matcher *ContainElementMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return Message(actual, "not to contain element matching", matcher.Element)
}

func (matcher *ContainElementMatcher) String() string {
	return Object(matcher.ContainElementMatcher, 0)
}
