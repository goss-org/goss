package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type OrMatcher struct {
	matchers.OrMatcher
}

func Or(ms ...types.GomegaMatcher) types.GomegaMatcher {
	return &OrMatcher{matchers.OrMatcher{Matchers: ms}}
}

func (matcher *OrMatcher) String() string {
	return Object(matcher.OrMatcher, 0)
}
