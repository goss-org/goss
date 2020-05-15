package matchers

import (
	"fmt"
	"strings"

	"github.com/onsi/gomega/types"
)

type WithSafeTransformMatcher struct {
	// input
	Transform Transformer // must be a function of one parameter that returns one value
	Matcher   types.GomegaMatcher

	// state
	transformedValue interface{}
	err              error
}

func WithSafeTransform(transform Transformer, matcher types.GomegaMatcher) *WithSafeTransformMatcher {

	return &WithSafeTransformMatcher{
		Transform: transform,
		Matcher:   matcher,
	}
}

func (m *WithSafeTransformMatcher) Match(actual interface{}) (bool, error) {
	var err error
	m.transformedValue, err = m.Transform.Transform(actual)
	if err != nil {
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

//func (m *WithSafeTransformMatcher) transformName() string {
//	fn := reflect.ValueOf(m.Transform)
//	n := runtime.FuncForPC(fn.Pointer()).Name()
//	ss := strings.Split(n, ".")
//	s := ss[len(ss)-1]
//	return s
//}
func (m *WithSafeTransformMatcher) String() string {
	return Object(m.Matcher, 1)
}

func getMatcherName(i interface{}) string {
	n := fmt.Sprintf("%#v", i)
	ss := strings.Split(n, ".")
	s := ss[len(ss)-1]
	return s
}
