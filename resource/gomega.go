package resource

import (
	"fmt"

	"github.com/goss-org/goss/matchers"
	"github.com/samber/lo"
)

func matcherToGomegaMatcher(matcher any) (matchers.GossMatcher, error) {
	// Default matchers
	switch x := matcher.(type) {
	case string:
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.Equal(x)), nil
	case float64, int:
		return matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically("eq", x)), nil
	case bool:
		return matchers.Equal(x), nil
	case []any:
		subMatchers, err := sliceToGomega(x, "")
		if err != nil {
			return nil, err
		}
		var interfaceSlice []any
		for _, d := range subMatchers {
			interfaceSlice = append(interfaceSlice, d)
		}
		return matchers.ContainElements(interfaceSlice...), nil
	}
	if matcher == nil {
		return nil, fmt.Errorf("Syntax Error: Missing required attribute")
	}
	matcherMap, ok := matcher.(map[string]any)
	if !ok {
		return nil, invalidArgSyntaxError("matcher", "map", matcher)
		//panic(fmt.Sprintf("Syntax Error: Unexpected matcher type: %T\n\n", matcher))
	}
	keys := lo.Keys(matcherMap)
	if len(keys) > 1 {
		return nil, fmt.Errorf("Syntax Error: Invalid matcher configuration. At a given nesting level, only one matcher is allowed. Found multiple matchers: %q", keys)
	}
	matchType := keys[0]
	value := matcherMap[matchType]
	switch matchType {
	case "equal":
		return matchers.Equal(value), nil
	case "have-prefix":
		v, isStr := value.(string)
		if !isStr {
			return nil, invalidArgSyntaxError("have-prefix", "string", value)
		}
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.HavePrefix(v)), nil
	case "have-suffix":
		v, isStr := value.(string)
		if !isStr {
			return nil, invalidArgSyntaxError("have-suffix", "string", value)
		}
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.HaveSuffix(v)), nil
	case "match-regexp":
		v, isStr := value.(string)
		if !isStr {
			return nil, invalidArgSyntaxError("match-regexp", "string", value)
		}
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.MatchRegexp(v)), nil
	case "contain-substring":
		v, isStr := value.(string)
		if !isStr {
			return nil, invalidArgSyntaxError("contain-substring", "string", value)

		}
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.ContainSubstring(v)), nil
	case "have-len":
		var v int
		switch val := value.(type) {
		case float64:
			v = int(val)
		case int:
			v = val
		default:
			return nil, invalidArgSyntaxError("have-len", "numeric", value)
		}
		return matchers.HaveLen(v), nil
	case "have-patterns":
		_, isArr := value.([]any)
		if !isArr {
			return nil, invalidArgSyntaxError("have-patterns", "array", value)

		}
		return matchers.WithSafeTransform(matchers.ToString{}, matchers.HavePatterns(value)), nil
	case "have-key":
		subMatcher, err := matcherToGomegaMatcher(value)
		if err != nil {
			return nil, err
		}
		return matchers.HaveKey(subMatcher), nil
	case "contain-element":
		switch value.(type) {
		case map[string]any, string, float64, int:
		default:
			return nil, invalidArgSyntaxError("contain-element", "matcher, string or numeric", value)

		}
		subMatcher, err := matcherToGomegaMatcher(value)
		if err != nil {
			return nil, err
		}
		return matchers.WithSafeTransform(matchers.ToArray{}, matchers.ContainElement(subMatcher)), nil
	case "contain-elements":
		subMatchers, err := sliceToGomega(value, "contains-elements")
		if err != nil {
			return nil, err
		}
		var interfaceSlice []any
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
		subMatchers, err := sliceToGomega(value, "consist-of")
		if err != nil {
			return nil, err
		}
		var interfaceSlice []any
		for _, d := range subMatchers {
			interfaceSlice = append(interfaceSlice, d)
		}
		return matchers.ConsistOf(interfaceSlice...), nil
	case "and":
		subMatchers, err := sliceToGomega(value, "and")
		if err != nil {
			return nil, err
		}
		return matchers.And(subMatchers...), nil
	case "or":
		subMatchers, err := sliceToGomega(value, "or")
		if err != nil {
			return nil, err
		}
		return matchers.Or(subMatchers...), nil
	case "gt", "ge", "lt", "le":
		return matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically(matchType, value)), nil

	case "semver-constraint":
		v, isStr := value.(string)
		if !isStr {
			return nil, invalidArgSyntaxError("semver-constraint", "string", value)

		}
		return matchers.BeSemverConstraint(v), nil
	case "gjson":
		var subMatchers []matchers.GossMatcher
		valueI, ok := value.(map[string]any)
		if !ok {
			return nil, invalidArgSyntaxError("gjson", "map", value)
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
		return nil, fmt.Errorf("Syntax Error: Unknown matcher: %s", matchType)

	}
}

func sliceToGomega(value any, name string) ([]matchers.GossMatcher, error) {
	valueI, ok := value.([]any)
	if !ok {
		return nil, invalidArgSyntaxError(name, "array", value)
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

func invalidArgSyntaxError(name, expected string, value any) error {
	return fmt.Errorf("Syntax Error: Invalid '%s' argument. Expected %s value, but received: %T: %q", name, expected, value, value)
}
