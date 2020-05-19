package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type HaveKeyMatcher struct {
	matchers.HaveKeyMatcher
}

func HaveKey(key interface{}) types.GomegaMatcher {
	return &HaveKeyMatcher{
		matchers.HaveKeyMatcher{
			Key: key,
		},
	}
}

func (matcher *HaveKeyMatcher) String() string {
	return Object(matcher.HaveKeyMatcher, 0)
}
