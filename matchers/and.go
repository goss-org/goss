package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type AndMatcher struct {
	matchers.AndMatcher
}

func And(ms ...types.GomegaMatcher) types.GomegaMatcher {
	return &AndMatcher{matchers.AndMatcher{Matchers: ms}}
}

//FIXME: Indentation is wrong
func (matcher *AndMatcher) String() string {
	return Object(matcher.AndMatcher, 0)
}
