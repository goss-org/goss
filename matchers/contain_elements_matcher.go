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
func (matcher *ContainElementsMatcher) FailureResult(actual interface{}) MatcherResult {
	missingElements := getUnexported(matcher, "missingElements")
	return MatcherResult{
		Actual:          actual,
		Message:         "to contain elements",
		Expected:        matcher.Elements,
		MissingElements: missingElements,
	}

}
func (matcher *ContainElementsMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to contain elements",
		Expected: matcher.Elements,
	}

}

func (matcher *ContainElementsMatcher) FailureMessage(actual interface{}) (message string) {
	//message = Message(actual, "to contain elements", matcher.Elements)
	//rs := reflect.ValueOf(matcher).Elem()
	////rs2 := reflect.New(rs.Type()).Elem()
	////rs2.Set(rs)
	//rf := rs.FieldByName("missingElements")
	//rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	rf := getUnexported(matcher, "missingElements")
	fmt.Println("wtf", rf)
	return fmt.Sprint("wtf2", rf)
	//return appendMissingElements(message, matcher.missingElements)
}

func (matcher *ContainElementsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return Message(actual, "not to contain elements", matcher.Elements)
}

func (matcher *ContainElementsMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["contain-elements"] = matcher.Elements
	return json.Marshal(j)
}

func (matcher *ContainElementsMatcher) String() string {
	return ""
	//return Object(matcher.GomegaContainElementsMatcher, 0)
}
