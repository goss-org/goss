package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type NotMatcher struct {
	matchers.NotMatcher
}

func Not(matcher types.GomegaMatcher) types.GomegaMatcher {
	return &NotMatcher{matchers.NotMatcher{Matcher: matcher}}
}

func (matcher *NotMatcher) String() string {
	return Object(matcher.NotMatcher, 0)
}
