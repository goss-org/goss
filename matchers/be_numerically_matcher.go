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

func (m *BeNumericallyMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  fmt.Sprintf("to be %s", m.Comparator),
		Expected: m.CompareTo[0],
	}
}

func (m *BeNumericallyMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  fmt.Sprintf("not to be %s", m.Comparator),
		Expected: m.CompareTo[0],
	}
}

func (m *BeNumericallyMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j[numericSymbolToStr[m.Comparator]] = m.CompareTo[0]
	return json.Marshal(j)
}

func (m *BeNumericallyMatcher) String() string {
	return Object(m.BeNumericallyMatcher, 0)
}

var numericSymbolToStr = map[string]string{
	">":  "gt",
	">=": "ge",
	"<":  "lt",
	"<=": "le",
}

// Stubs to match omegaMatcher
func (m *BeNumericallyMatcher) FailureMessage(_ interface{}) (message string) {
	return ""
}

// Stubs to match omegaMatcher
func (m *BeNumericallyMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return ""
}
