package matchers

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/onsi/gomega/format"
)

type GossMatcher interface {
	Match(actual interface{}) (success bool, err error)
	FailureResult(actual interface{}) MatcherResult
	NegatedFailureResult(actual interface{}) MatcherResult
	fmt.Stringer
}

type MatcherResult struct {
	Actual           interface{}
	Message          string
	Expected         interface{}
	MissingElements  interface{}
	ExtraElements    interface{}
	TransformerChain []Transformer
	Successful       bool
}

func (m MatcherResult) String() string {
	if m.Successful {
		return fmt.Sprintf("matches expectation: %s", m.Expected)
	}
	return format.Message(m.Actual, m.Message, m.Expected)
}

func getUnexported(i interface{}, field string) interface{} {
	rs := reflect.ValueOf(i).Elem()
	rf := rs.FieldByName(field)
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	return fmt.Sprint(rf)
}
