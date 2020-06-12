package matchers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/sanity-io/litter"
)

type WithSafeTransformMatcher struct {
	// input
	Transform Transformer // must be a function of one parameter that returns one value
	Matcher   GossMatcher

	// state
	transformedValue interface{}
	wasTransformed   bool
	err              error
}

func WithSafeTransform(transform Transformer, matcher GossMatcher) *WithSafeTransformMatcher {

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

func (m *WithSafeTransformMatcher) FailureResult(_ interface{}) MatcherResult {
	tchain, matcher := m.getTransformerChainAndMatcher()
	result := matcher.FailureResult(m.transformedValue)
	result.TransformerChain = tchain
	return result
}
func (m *WithSafeTransformMatcher) NegatedFailureResult(_ interface{}) MatcherResult {
	tchain, matcher := m.getTransformerChainAndMatcher()
	result := matcher.NegatedFailureResult(m.transformedValue)
	result.TransformerChain = tchain
	return result
}

// Stubs to match omegaMatcher
func (m *WithSafeTransformMatcher) FailureMessage(_ interface{}) (message string) {
	return ""
}

// Stubs to match omegaMatcher
func (m *WithSafeTransformMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return ""
}

func (m *WithSafeTransformMatcher) getTransformerChainAndMatcher() (tchain []Transformer, matcher GossMatcher) {
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

func (m *WithSafeTransformMatcher) MarshalJSON() ([]byte, error) {
	tchain, matcher := m.getTransformerChainAndMatcher()
	//if len(tchain) == 0 {
	//	return json.Marshal(matcher)
	//}
	if len(tchain) == 0 || true {
		return json.Marshal(matcher)
	}
	j := make(map[string]interface{})
	j["matcher"] = matcher
	//j["transform-chain"] = tchain
	sq := litter.Options{Compact: true, StripPackageNames: true}
	ss := make([]string, len(tchain))
	for i, v := range tchain {
		ss[i] = sq.Sdump(v)
	}
	j["transform-chain"] = ss
	return json.Marshal(j)
	////fmt.Println("wtf5", m.String())
	//return json.Marshal(m)
	//return []byte(m.String()), nil
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
