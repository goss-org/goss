package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type ContainElementsMatcher struct {
	matchers.ContainElementsMatcher
}

func ContainElements(elements ...interface{}) GossMatcher {
	return &ContainElementsMatcher{
		matchers.ContainElementsMatcher{
			Elements: elements,
		},
	}
}
func (m *ContainElementsMatcher) FailureResult(actual interface{}) MatcherResult {
	missingElements := getUnexported(m, "missingElements")
	return MatcherResult{
		Actual:          actual,
		Message:         "to contain elements matching",
		Expected:        m.Elements,
		MissingElements: missingElements,
	}

}
func (m *ContainElementsMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to contain elements matching",
		Expected: m.Elements,
	}

}

func (m *ContainElementsMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["contain-elements"] = m.Elements
	return json.Marshal(j)
}
