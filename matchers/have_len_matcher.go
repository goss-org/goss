package matchers

import (
	"fmt"

	"github.com/onsi/gomega/matchers"
)

type HaveLenMatcher struct {
	matchers.HaveLenMatcher
}

func HaveLen(count int) GossMatcher {
	return &HaveLenMatcher{
		matchers.HaveLenMatcher{
			Count: count,
		},
	}
}

func (matcher *HaveLenMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have length",
		Expected: matcher.Count,
	}
}

func (matcher *HaveLenMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have length",
		Expected: matcher.Count,
	}
}

func (matcher *HaveLenMatcher) String() string {
	return fmt.Sprintf("HaveLen{Count:%d}", matcher.Count)
}

//func (matcher *HaveLenMatcher) String() string {
//	n := fmt.Sprintf("%#v", matcher.HaveLenMatcher)
//	ss := strings.Split(n, ".")
//	s := ss[len(ss)-1]
//	return s
//}
