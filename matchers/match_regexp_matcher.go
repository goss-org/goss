package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type MatchRegexpMatcher struct {
	matchers.MatchRegexpMatcher
}

func MatchRegexp(regexp string, args ...interface{}) types.GomegaMatcher {
	return &MatchRegexpMatcher{
		matchers.MatchRegexpMatcher{
			Regexp: regexp,
			Args:   args,
		},
	}
}

func (matcher *MatchRegexpMatcher) String() string {
	return Object(matcher.MatchRegexpMatcher, 0)
}
