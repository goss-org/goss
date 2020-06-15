package matchers

import (
	"encoding/json"
	"reflect"
	"unsafe"
)

type GossMatcher interface {
	Match(actual interface{}) (success bool, err error)
	FailureResult(actual interface{}) MatcherResult
	NegatedFailureResult(actual interface{}) MatcherResult
	json.Marshaler
}

type MatcherResult struct {
	Actual             interface{}
	Message            string
	Expected           interface{}
	MissingElements    interface{}
	ExtraElements      interface{}
	TransformerChain   []Transformer
	UntransformedValue interface{}
}

func getUnexported(i interface{}, field string) interface{} {
	rs := reflect.ValueOf(i).Elem()
	rf := rs.FieldByName(field)
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	return rf.Interface()
}
