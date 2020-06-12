package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega/matchers"
)

type BeNumericallyMatcher struct {
	matchers.BeNumericallyMatcher
}

func BeNumerically(comparator string, compareTo ...interface{}) GossMatcher {
	return &BeNumericallyMatcher{
		matchers.BeNumericallyMatcher{
			Comparator: comparator,
			CompareTo:  compareTo,
		},
	}
}

func (matcher *BeNumericallyMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  fmt.Sprintf("to be %s", numericSymbolToStr[matcher.Comparator]),
		Expected: matcher.CompareTo[0],
	}
}

func (matcher *BeNumericallyMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  fmt.Sprintf("not to be %s", numericSymbolToStr[matcher.Comparator]),
		Expected: matcher.CompareTo[0],
	}
}

func (matcher *BeNumericallyMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j[numericSymbolToStr[matcher.Comparator]] = matcher.CompareTo[0]
	return json.Marshal(j)
}

func (matcher *BeNumericallyMatcher) String() string {
	return Object(matcher.BeNumericallyMatcher, 0)
}

var numericSymbolToStr = map[string]string{
	">":  "gt",
	">=": "ge",
	"<":  "lt",
	"<=": "le",
}
