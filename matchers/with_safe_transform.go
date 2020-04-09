package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/types"
)

type WithSafeTransformMatcher struct {
	// input
	Transform interface{} // must be a function of one parameter that returns one value
	Matcher   types.GomegaMatcher

	// cached value
	transformArgType reflect.Type

	// state
	transformedValue interface{}
	err              error
}

func WithSafeTransform(transform interface{}, matcher types.GomegaMatcher) *WithSafeTransformMatcher {
	if transform == nil {
		panic("transform function cannot be nil")
	}
	txType := reflect.TypeOf(transform)
	if txType.NumIn() != 1 {
		panic("transform function must have 1 argument")
	}
	if txType.NumOut() != 2 {
		panic("transform function must have 2 return value")
	}

	return &WithSafeTransformMatcher{
		Transform:        transform,
		Matcher:          matcher,
		transformArgType: reflect.TypeOf(transform).In(0),
	}
}

func (m *WithSafeTransformMatcher) Match(actual interface{}) (bool, error) {
	// return error if actual's type is incompatible with Transform function's argument type
	actualType := reflect.TypeOf(actual)
	if !actualType.AssignableTo(m.transformArgType) {
		return false, fmt.Errorf("Transform function expects '%s' but we have '%s'", m.transformArgType, actualType)
	}

	// call the Transform function with `actual`
	fn := reflect.ValueOf(m.Transform)
	result := fn.Call([]reflect.Value{reflect.ValueOf(actual)})
	m.transformedValue = result[0].Interface() // expect exactly one value
	if err, ok := result[1].Interface().(error); ok && err != nil {
		return false, err
	}

	return m.Matcher.Match(m.transformedValue)
}

func (m *WithSafeTransformMatcher) FailureMessage(_ interface{}) (message string) {
	return m.Matcher.FailureMessage(m.transformedValue)
}

func (m *WithSafeTransformMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return m.Matcher.NegatedFailureMessage(m.transformedValue)
}
