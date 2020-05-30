package matchers

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type ContainElementsMatcher struct {
	matchers.ContainElementsMatcher
}

func ContainElements(elements ...interface{}) types.GomegaMatcher {
	return &ContainElementsMatcher{
		matchers.ContainElementsMatcher{
			Elements: elements,
		},
	}
}

func (matcher *ContainElementsMatcher) FailureMessage(actual interface{}) (message string) {
	message = Message(actual, "to contain elements", matcher.Elements)
	rs := reflect.ValueOf(matcher).Elem()
	//rs2 := reflect.New(rs.Type()).Elem()
	//rs2.Set(rs)
	rf := rs.FieldByName("missingElements")
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	fmt.Println("wtf", rf)
	return ""
	//return appendMissingElements(message, matcher.missingElements)
}

func (matcher *ContainElementsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return Message(actual, "not to contain elements", matcher.Elements)
}

func (matcher *ContainElementsMatcher) String() string {
	return ""
	//return Object(matcher.GomegaContainElementsMatcher, 0)
}
