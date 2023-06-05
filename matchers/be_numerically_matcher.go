package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega/matchers"
)

type BeNumericallyMatcher struct {
	fakeOmegaMatcher
	Comparator string
	CompareTo  []interface{}
	//matchers.BeNumericallyMatcher
}

func BeNumerically(comparator string, compareTo ...interface{}) GossMatcher {
	return &BeNumericallyMatcher{
		Comparator: comparator,
		CompareTo:  compareTo,
	}
}
func (m *BeNumericallyMatcher) Match(actual interface{}) (success bool, err error) {
	comparator, err := strToSymbol(m.Comparator)
	if err != nil {
		return false, err
	}
	matcher := &matchers.BeNumericallyMatcher{
		Comparator: comparator,
		CompareTo:  m.CompareTo,
	}
	return matcher.Match(actual)
}

func (m *BeNumericallyMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  fmt.Sprintf("to be numerically %s", m.Comparator),
		Expected: m.CompareTo[0],
	}
}

func (m *BeNumericallyMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  fmt.Sprintf("not to be numerically %s", m.Comparator),
		Expected: m.CompareTo[0],
	}
}

func (m *BeNumericallyMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j[m.Comparator] = m.CompareTo[0]
	return json.Marshal(j)
}

func strToSymbol(s string) (string, error) {
	comparator, ok := map[string]string{
		"gt": ">",
		"ge": ">=",
		"lt": "<",
		"le": "<=",
		"eq": "==",
	}[s]
	if !ok {
		return "", fmt.Errorf("Unknown comparator: %s", s)
	}
	return comparator, nil
}
