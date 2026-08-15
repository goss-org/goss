package matchers

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type WithSafeTransformMatcher struct {
	fakeOmegaMatcher

	// input
	Transform Transformer // must be a function of one parameter that returns one value
	Matcher   GossMatcher

	// state
	transformedValue any
	wasTransformed   bool
}

func WithSafeTransform(transform Transformer, matcher GossMatcher) GossMatcher {
	return &WithSafeTransformMatcher{
		Transform: transform,
		Matcher:   matcher,
	}
}

func (m *WithSafeTransformMatcher) Match(actual any) (bool, error) {
	var err error
	//log.Printf("%#v: input: %v", m.Transform, actual)
	m.transformedValue, err = m.Transform.Transform(actual)
	if !reflect.DeepEqual(actual, m.transformedValue) {
		m.wasTransformed = true
	}
	if err != nil {
		return false, fmt.Errorf("%#v: %w", m.Transform, err)
	}
	//log.Printf("%#v: output: %v", m.Transform, m.transformedValue)
	return m.Matcher.Match(m.transformedValue)
}

func (m *WithSafeTransformMatcher) FailureResult(actual any) MatcherResult {
	tchain, matcher, tvalue := m.getTransformerChainAndMatcher()
	result := matcher.FailureResult(tvalue)
	result.TransformerChain = tchain
	result.UntransformedValue = actual
	return result
}

func (m *WithSafeTransformMatcher) NegatedFailureResult(actual any) MatcherResult {
	tchain, matcher, tvalue := m.getTransformerChainAndMatcher()
	result := matcher.NegatedFailureResult(tvalue)
	result.TransformerChain = tchain
	result.UntransformedValue = actual
	return result
}

func (m *WithSafeTransformMatcher) getTransformerChainAndMatcher() ([]Transformer, GossMatcher, any) {
	var matcher GossMatcher = m
	tchain := make([]Transformer, 0)
	tvalue := m.transformedValue
L:
	for {
		switch v := matcher.(type) {
		case *WithSafeTransformMatcher:
			matcher = v.Matcher
			tvalue = v.transformedValue
			if v.wasTransformed {
				tchain = append(tchain, v.Transform)
			}
		default:
			break L
		}
	}
	return tchain, matcher, tvalue
}

func (m *WithSafeTransformMatcher) MarshalJSON() ([]byte, error) {
	_, matcher, _ := m.getTransformerChainAndMatcher()
	return json.Marshal(matcher)
}
