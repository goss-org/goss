package matchers

import (
	"github.com/onsi/gomega/types"
)

type ConsistOfMatcher struct {
	GomegaConsistOfMatcher
}

func ConsistOf(elements ...interface{}) types.GomegaMatcher {
	return &ConsistOfMatcher{
		GomegaConsistOfMatcher{
			Elements: elements,
		},
	}
}

func (matcher *ConsistOfMatcher) String() string {
	return Object(matcher.GomegaConsistOfMatcher, 0)
}
