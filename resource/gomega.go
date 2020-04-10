package resource

import (
	"fmt"
	"sort"

	"github.com/aelsabbahy/goss/matchers"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func matcherToGomegaMatcher(matcher interface{}) (types.GomegaMatcher, error) {
	switch x := matcher.(type) {
	case string:
		return gomega.WithTransform(matchers.ToString, gomega.Equal(x)), nil
	case int, bool, float64:
		return gomega.Equal(x), nil
	case []interface{}:
		var matchers []types.GomegaMatcher
		for _, valueI := range x {
			if subMatcher, ok := valueI.(types.GomegaMatcher); ok {
				matchers = append(matchers, subMatcher)
			} else {
				subMatcher2, err := matcherToGomegaMatcher(valueI)
				if err != nil {
					return nil, err
				}
				matchers = append(matchers, gomega.ContainElement(subMatcher2))
			}
		}
		return gomega.And(matchers...), nil
	}
	matcher = sanitizeExpectedValue(matcher)
	if matcher == nil {
		return nil, fmt.Errorf("Missing Required Attribute")
	}
	matcherMap, ok := matcher.(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("Unexpected matcher type: %T\n\n", matcher))
	}
	var matchType string
	var value interface{}
	for matchType, value = range matcherMap {
		break
	}
	switch matchType {
	case "have-prefix":
		return gomega.WithTransform(matchers.ToString, gomega.HavePrefix(value.(string))), nil
	case "have-suffix":
		return gomega.WithTransform(matchers.ToString, gomega.HaveSuffix(value.(string))), nil
	case "match-regexp":
		return gomega.WithTransform(matchers.ToString, gomega.MatchRegexp(value.(string))), nil
	case "have-len":
		value = sanitizeExpectedValue(value)
		return gomega.HaveLen(value.(int)), nil
	case "have-key-with-value":
		subMatchers, err := mapToGomega(value)
		if err != nil {
			return nil, err
		}
		for key, val := range subMatchers {
			if val == nil {
				fmt.Printf("%d is nil", key)
			}
		}
		return gomega.And(subMatchers...), nil
	case "have-key":
		subMatcher, err := matcherToGomegaMatcher(value)
		if err != nil {
			return nil, err
		}
		return gomega.HaveKey(subMatcher), nil
	case "contain-element":
		subMatcher, err := matcherToGomegaMatcher(value)
		if err != nil {
			return nil, err
		}
		return gomega.ContainElement(subMatcher), nil
	case "contain-elements":
		subMatchers, err := sliceToGomega(value)
		if err != nil {
			return nil, err
		}
		var interfaceSlice []interface{}
		for _, d := range subMatchers {
			interfaceSlice = append(interfaceSlice, d)
		}
		return gomega.ContainElements(interfaceSlice...), nil
	case "not":
		subMatcher, err := matcherToGomegaMatcher(value)
		if err != nil {
			return nil, err
		}
		return gomega.Not(subMatcher), nil
	case "consist-of":
		subMatchers, err := sliceToGomega(value)
		if err != nil {
			return nil, err
		}
		var interfaceSlice []interface{}
		for _, d := range subMatchers {
			interfaceSlice = append(interfaceSlice, d)
		}
		return gomega.ConsistOf(interfaceSlice...), nil
	case "and":
		subMatchers, err := sliceToGomega(value)
		if err != nil {
			return nil, err
		}
		return gomega.And(subMatchers...), nil
	case "or":
		subMatchers, err := sliceToGomega(value)
		if err != nil {
			return nil, err
		}
		return gomega.Or(subMatchers...), nil
	case "gt", "ge", "lt", "le":
		// Golang json escapes '>', '<' symbols, so we use 'gt', 'le' instead
		comparator := map[string]string{
			"gt": ">",
			"ge": ">=",
			"lt": "<",
			"le": "<=",
		}[matchType]
		return matchers.WithSafeTransform(matchers.ToFloat64, gomega.BeNumerically(comparator, value)), nil

	case "semver-constraint":
		return matchers.BeSemverConstraint(value.(string)), nil
	default:
		return nil, fmt.Errorf("Unknown matcher: %s", matchType)

	}
}

func mapToGomega(value interface{}) (subMatchers []types.GomegaMatcher, err error) {
	valueI, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Matcher expected map, got: %t", value)
	}

	// Get keys
	keys := []string{}
	for key, _ := range valueI {
		keys = append(keys, key)
	}
	// Iterate through keys in a deterministic way, since ranging over a map
	// does not guarantee order
	sort.Strings(keys)
	for _, key := range keys {
		val := valueI[key]
		val, err = matcherToGomegaMatcher(val)
		if err != nil {
			return
		}

		subMatcher := gomega.HaveKeyWithValue(key, val)
		subMatchers = append(subMatchers, subMatcher)
	}
	return
}

func sliceToGomega(value interface{}) ([]types.GomegaMatcher, error) {
	valueI, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Matcher expected array, got: %t", value)
	}
	var subMatchers []types.GomegaMatcher
	for _, v := range valueI {
		subMatcher, err := matcherToGomegaMatcher(v)
		if err != nil {
			return nil, err
		}
		subMatchers = append(subMatchers, subMatcher)
	}
	return subMatchers, nil
}

// Normalize expectedValue so json and yaml are the same
func sanitizeExpectedValue(i interface{}) interface{} {
	if e, ok := i.(float64); ok {
		return int(e)
	}
	if e, ok := i.(map[interface{}]interface{}); ok {
		out := make(map[string]interface{})
		for k, v := range e {
			ks, ok := k.(string)
			if !ok {
				panic(fmt.Sprintf("Matcher key type not string: %T\n\n", k))
			}
			out[ks] = sanitizeExpectedValue(v)
		}
		return out
	}
	return i
}
