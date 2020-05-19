package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type BeNumericallyMatcher struct {
	matchers.BeNumericallyMatcher
}

func BeNumerically(comparator string, compareTo ...interface{}) types.GomegaMatcher {
	return &BeNumericallyMatcher{
		matchers.BeNumericallyMatcher{
			Comparator: comparator,
			CompareTo:  compareTo,
		},
	}
}

func (matcher *BeNumericallyMatcher) String() string {
	return Object(matcher.BeNumericallyMatcher, 0)
}
