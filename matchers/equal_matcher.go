package matchers

import (
	"fmt"

	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type EqualMatcher struct {
	matchers.EqualMatcher
}

func Equal(element interface{}) types.GomegaMatcher {
	return &EqualMatcher{
		matchers.EqualMatcher{
			Expected: element,
		},
	}
}

func (matcher *EqualMatcher) String() string {
	return fmt.Sprint(matcher.Expected)
}
