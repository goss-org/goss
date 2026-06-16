package matchers

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/matchers"
	"github.com/samber/lo"
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

func (m *ContainElementsMatcher) Match(actual any) (success bool, err error) {
	if !isArrayOrSlice(actual) && !isMap(actual) {
		return false, fmt.Errorf("ContainElements matcher expects an array/slice/map.  Got:\n%s", format.Object(actual, 1))
	}
	return m.ContainElementsMatcher.Match(actual)
}

func (m *ContainElementsMatcher) FailureResult(actual interface{}) MatcherResult {
	missingElements := getUnexported(m, "missingElements")
	missingEl, ok := missingElements.([]interface{})
	var foundElements any
	if ok {
		foundElements, _ = lo.Difference(m.Elements, missingEl)
	}
	return MatcherResult{
		Actual:          actual,
		Message:         "to contain elements matching",
		Expected:        m.Elements,
		MissingElements: missingElements,
		FoundElements:   foundElements,
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

func isMap(a any) bool {
	if a == nil {
		return false
	}
	return reflect.TypeOf(a).Kind() == reflect.Map
}

func isArrayOrSlice(a any) bool {
	if a == nil {
		return false
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}
