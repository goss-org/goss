package resource

import (
	"fmt"
	"log"

	"github.com/aelsabbahy/goss/matchers"
)

func matcherToGomegaMatcher(matcher interface{}) (matchers.GossMatcher, error) {
	switch x := matcher.(type) {
	case string:
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.Equal(x)), nil
	case float64, int:
		return matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically("==", x)), nil
	case bool:
		return matchers.Equal(x), nil
	case []interface{}:
		subMatchers, err := sliceToGomega(x)
		if err != nil {
			return nil, err
		}
		var interfaceSlice []interface{}
		for _, d := range subMatchers {
			interfaceSlice = append(interfaceSlice, d)
		}
		return matchers.ContainElements(interfaceSlice...), nil
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
	case "equal":
		return matchers.Equal(value), nil
	case "have-prefix":
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.HavePrefix(value.(string))), nil
	case "have-suffix":
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.HaveSuffix(value.(string))), nil
	case "match-regexp":
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.MatchRegexp(value.(string))), nil
	case "contain-substring":
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.ContainSubstring(value.(string))), nil
	case "have-len":
		return matchers.HaveLen(int(value.(float64))), nil
	case "have-patterns":
		_, isArr := value.([]interface{})
		if !isArr {
			return nil, fmt.Errorf("have-patterns: incorrect expectation type, expected array, got: %t", value)
		}
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.HavePatterns(value)), nil
	case "have-key":
		subMatcher, err := matcherToGomegaMatcher(value)
		if err != nil {
			return nil, err
		}
		return matchers.HaveKey(subMatcher), nil
	case "contain-element":
		_, isArr := value.([]interface{})
		if isArr {
			return nil, fmt.Errorf("contain-element: incorrect expectation type, expected matcher or value, got: %t", value)
		}
		subMatcher, err := matcherToGomegaMatcher(value)
		log.Printf("output: %#v", value)
		if err != nil {
			return nil, err
		}
		return matchers.WithSafeTransform(matchers.ToArray{}, matchers.ContainElement(subMatcher)), nil
	case "contain-elements":
		subMatchers, err := sliceToGomega(value)
		if err != nil {
			return nil, err
		}
		var interfaceSlice []interface{}
		for _, d := range subMatchers {
			interfaceSlice = append(interfaceSlice, d)
		}
		return matchers.WithSafeTransform(matchers.ToArray{}, matchers.ContainElements(interfaceSlice...)), nil
	case "not":
		subMatcher, err := matcherToGomegaMatcher(value)
		if err != nil {
			return nil, err
		}
		return matchers.Not(subMatcher), nil
	case "consist-of":
		subMatchers, err := sliceToGomega(value)
		if err != nil {
			return nil, err
		}
		var interfaceSlice []interface{}
		for _, d := range subMatchers {
			interfaceSlice = append(interfaceSlice, d)
		}
		return matchers.ConsistOf(interfaceSlice...), nil
	case "and":
		subMatchers, err := sliceToGomega(value)
		if err != nil {
			return nil, err
		}
		return matchers.And(subMatchers...), nil
	case "or":
		subMatchers, err := sliceToGomega(value)
		if err != nil {
			return nil, err
		}
		return matchers.Or(subMatchers...), nil
	case "gt", "ge", "lt", "le":
		// Golang json escapes '>', '<' symbols, so we use 'gt', 'le' instead
		comparator := map[string]string{
			"gt": ">",
			"ge": ">=",
			"lt": "<",
			"le": "<=",
		}[matchType]
		return matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically(comparator, value)), nil

	case "semver-constraint":
		return matchers.BeSemverConstraint(value.(string)), nil
	case "gjson":
		var subMatchers []matchers.GossMatcher
		valueI, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Matcher expected map, got: %t", value)
		}
		for key, val := range valueI {
			subMatcher, err := matcherToGomegaMatcher(val)
			if err != nil {
				return nil, err
			}
			subMatchers = append(subMatchers, matchers.WithSafeTransform(matchers.Gjson{Path: key}, subMatcher))

		}
		return matchers.And(subMatchers...), nil
	default:
		return nil, fmt.Errorf("Unknown matcher: %s", matchType)

	}
}

func sliceToGomega(value interface{}) ([]matchers.GossMatcher, error) {
	valueI, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Matcher expected array, got: %t", value)
	}
	var subMatchers []matchers.GossMatcher
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
	if e, ok := i.(int); ok {
		return float64(e)
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
