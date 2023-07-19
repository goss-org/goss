package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type ContainElementMatcher struct {
	matchers.ContainElementMatcher
}

func ContainElement(element interface{}) GossMatcher {
	return &ContainElementMatcher{
		matchers.ContainElementMatcher{
			Element: element,
		},
	}
}

func (m *ContainElementMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to contain element matching",
		Expected: m.Element,
	}
}

func (m *ContainElementMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to contain element matching",
		Expected: m.Element,
	}
}

func (m *ContainElementMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["contain-element"] = m.Element
	return json.Marshal(j)
}
