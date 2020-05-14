package matchers

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

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
	//fmt.Printf("%v\n", m.transformedValue)
	//fmt.Printf("%T\n", m.transformedValue)
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

func (m *WithSafeTransformMatcher) transformName() string {
	fn := reflect.ValueOf(m.Transform)
	n := runtime.FuncForPC(fn.Pointer()).Name()
	ss := strings.Split(n, ".")
	s := ss[len(ss)-1]
	return s
}
func (m *WithSafeTransformMatcher) String() string {
	//return fmt.Sprintf("Matcher: %#v\nTransform: %s", m.Matcher, m.transformName())
	//return fmt.Sprintf("%s{} | %#v", m.transformName(), m.Matcher)
	//return fmt.Sprintf("TransformMatcher{Transform:%s{}, Matcher:%s}", m.transformName(), getMatcherName(m.Matcher))
	return Object(m.Matcher, 1)
}

func getMatcherName(i interface{}) string {
	n := fmt.Sprintf("%#v", i)
	ss := strings.Split(n, ".")
	s := ss[len(ss)-1]
	return s
}
