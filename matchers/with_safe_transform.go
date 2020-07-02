package matchers

import (
	"encoding/json"
	"reflect"
)

type WithSafeTransformMatcher struct {
	fakeOmegaMatcher

	// input
	Transform Transformer // must be a function of one parameter that returns one value
	Matcher   GossMatcher

	// state
	transformedValue interface{}
	wasTransformed   bool
	err              error
}

func WithSafeTransform(transform Transformer, matcher GossMatcher) GossMatcher {

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

func (m *WithSafeTransformMatcher) FailureResult(actual interface{}) MatcherResult {
	tchain, matcher := m.getTransformerChainAndMatcher()
	result := matcher.FailureResult(m.transformedValue)
	result.TransformerChain = tchain
	result.UntransformedValue = actual
	return result
}
func (m *WithSafeTransformMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	tchain, matcher := m.getTransformerChainAndMatcher()
	result := matcher.NegatedFailureResult(m.transformedValue)
	result.TransformerChain = tchain
	result.UntransformedValue = actual
	return result
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

func (m *WithSafeTransformMatcher) MarshalJSON() ([]byte, error) {
	_, matcher := m.getTransformerChainAndMatcher()
	return json.Marshal(matcher)
}
