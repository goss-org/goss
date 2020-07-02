package matchers

import (
	"encoding/json"
	"fmt"

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
		Message:         "to contain elements",
		Expected:        m.Elements,
		MissingElements: missingElements,
	}

}
func (m *ContainElementsMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to contain elements",
		Expected: m.Elements,
	}

}

func (m *ContainElementsMatcher) FailureMessage(actual interface{}) (message string) {
	//message = Message(actual, "to contain elements", matcher.Elements)
	//rs := reflect.ValueOf(matcher).Elem()
	////rs2 := reflect.New(rs.Type()).Elem()
	////rs2.Set(rs)
	//rf := rs.FieldByName("missingElements")
	//rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	rf := getUnexported(m, "missingElements")
	fmt.Println("wtf", rf)
	return fmt.Sprint("wtf2", rf)
	//return appendMissingElements(message, matcher.missingElements)
}

func (m *ContainElementsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return Message(actual, "not to contain elements", m.Elements)
}

func (m *ContainElementsMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["contain-elements"] = m.Elements
	return json.Marshal(j)
}

func (m *ContainElementsMatcher) String() string {
	return ""
	//return Object(m.GomegaContainElementsMatcher, 0)
}
