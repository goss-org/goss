package matchers

import (
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type ContainSubstringMatcher struct {
	matchers.ContainSubstringMatcher
}

func ContainSubstring(substr string, args ...interface{}) types.GomegaMatcher {
	return &ContainSubstringMatcher{
		matchers.ContainSubstringMatcher{
			Substr: substr,
			Args:   args,
		},
	}
}

func (matcher *ContainSubstringMatcher) String() string {
	return format.Object(matcher.ContainSubstringMatcher, 0)
}
