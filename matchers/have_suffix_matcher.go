package matchers

import (
	"github.com/onsi/gomega/matchers"
)

type HaveSuffixMatcher struct {
	matchers.HaveSuffixMatcher
}

func HaveSuffix(prefix string, args ...interface{}) GossMatcher {
	return &HaveSuffixMatcher{
		matchers.HaveSuffixMatcher{
			Suffix: prefix,
			Args:   args,
		},
	}
}

func (matcher *HaveSuffixMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have suffix",
		Expected: matcher.Suffix,
	}
}

func (matcher *HaveSuffixMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have suffix",
		Expected: matcher.Suffix,
	}
}

func (matcher *HaveSuffixMatcher) String() string {
	return Object(matcher.HaveSuffixMatcher, 0)
}

//func (matcher *HaveSuffixMatcher) String() string {
//	return fmt.Sprintf("%s{Suffix: %s}", getObjectTypeName(matcher), matcher.Prefix)
//}
//
//func getObjectTypeName(m interface{}) string {
//	return strings.Split(reflect.TypeOf(m).String(), ".")[1]
//
//}
