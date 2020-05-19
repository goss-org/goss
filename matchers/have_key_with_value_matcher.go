package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type HaveKeyWithValueMatcher struct {
	matchers.HaveKeyWithValueMatcher
}

func HaveKeyWithValue(key interface{}, value interface{}) types.GomegaMatcher {
	return &HaveKeyWithValueMatcher{
		matchers.HaveKeyWithValueMatcher{
			Key:   key,
			Value: value,
		},
	}
}

func (matcher *HaveKeyWithValueMatcher) String() string {
	return Object(matcher.HaveKeyWithValueMatcher, 0)
}
