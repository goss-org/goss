package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega/matchers"
)

type BeNumericallyMatcher struct {
	fakeOmegaMatcher
	Comparator string
	CompareTo  []any
}

func BeNumerically(comparator string, compareTo ...any) GossMatcher {
	return &BeNumericallyMatcher{
		Comparator: comparator,
		CompareTo:  compareTo,
	}
}

func (m *BeNumericallyMatcher) Match(actual any) (bool, error) {
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

func (m *BeNumericallyMatcher) FailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to be numerically " + m.Comparator,
		Expected: m.CompareTo[0],
	}
}

func (m *BeNumericallyMatcher) NegatedFailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to be numerically " + m.Comparator,
		Expected: m.CompareTo[0],
	}
}

func (m *BeNumericallyMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]any)
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
