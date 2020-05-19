package matchers

import (
	"github.com/onsi/gomega/types"
)

type ContainElementsMatcher struct {
	GomegaContainElementsMatcher
}

func ContainElements(elements ...interface{}) types.GomegaMatcher {
	return &ContainElementsMatcher{
		GomegaContainElementsMatcher{
			Elements: elements,
		},
	}
}

func (matcher *ContainElementsMatcher) FailureMessage(actual interface{}) (message string) {
	message = Message(actual, "to contain elements", matcher.Elements)
	return appendMissingElements(message, matcher.MissingElements)
}

func (matcher *ContainElementsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return Message(actual, "not to contain elements", matcher.Elements)
}

func (matcher *ContainElementsMatcher) String() string {
	return Object(matcher.GomegaContainElementsMatcher, 0)
}
