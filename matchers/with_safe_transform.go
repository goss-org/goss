package matchers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/onsi/gomega/types"
	"github.com/sanity-io/litter"
)

type WithSafeTransformMatcher struct {
	// input
	Transform Transformer // must be a function of one parameter that returns one value
	Matcher   types.GomegaMatcher

	// state
	transformedValue interface{}
	wasTransformed   bool
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
	if !reflect.DeepEqual(actual, m.transformedValue) {
		m.wasTransformed = true
	}
	return m.Matcher.Match(m.transformedValue)
}

func (m *WithSafeTransformMatcher) FailureMessage(_ interface{}) (message string) {
	tchain, matcher := m.getTransformerChainAndMatcher()
	message = matcher.FailureMessage(m.transformedValue)
	return appendTransformMessage(message, tchain)
}

func (m *WithSafeTransformMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	tchain, matcher := m.getTransformerChainAndMatcher()
	message = matcher.NegatedFailureMessage(m.transformedValue)
	return appendTransformMessage(message, tchain)
}

func (m *WithSafeTransformMatcher) getTransformerChainAndMatcher() (tchain []Transformer, matcher types.GomegaMatcher) {
	matcher = m
L:
	for {
		switch v := matcher.(type) {
		case *WithSafeTransformMatcher:
			matcher = v.Matcher
			if v.wasTransformed {
				tchain = append(tchain, v.Transform)
			}
		default:
			break L

		}
	}
	return tchain, matcher

}
func appendTransformMessage(message string, tchain []Transformer) string {
	if len(tchain) == 0 {
		return message
	}
	var s string
	//for _, t := range tchain {
	//	s += fmt.Sprintf("%s\n", strings.TrimRight(format.Object(t, 1), " "))
	//}
	//s = litter.Sdump(tchain)
	sq := litter.Options{Compact: true, StripPackageNames: true}
	s += Indent
	for _, t := range tchain {
		s += " | " + sq.Sdump(t)
	}
	return fmt.Sprintf("%s\nwith transform chain\n%s", message,
		s)
}

func (m *WithSafeTransformMatcher) String() string {
	tchain, matcher := m.getTransformerChainAndMatcher()
	if len(tchain) == 0 {
		return fmt.Sprintf("%v", matcher)
	}
	sq := litter.Options{Compact: true}
	ss := make([]string, len(tchain))
	for i, v := range tchain {
		ss[i] = sq.Sdump(v)
	}
	ss = append(ss, fmt.Sprintf("%#v", matcher))
	return strings.Join(ss, "|")
	//return Object(matcher, 0)
}
