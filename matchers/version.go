package matchers

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/go-version"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func BeVersion(comparator string, compareTo interface{}) types.GomegaMatcher {
	return &BeVersionMatcher{
		Comparator: comparator,
		CompareTo:  compareTo,
	}
}

type BeVersionMatcher struct {
	Comparator string
	CompareTo  interface{}
}

func (matcher *BeVersionMatcher) Match(actual interface{}) (success bool, err error) {
	comparator, ok := map[string]func(*version.Version, *version.Version) bool{
		"==": func(v1 *version.Version, v2 *version.Version) bool { return v1.Equal(v2) },
		">":  func(v1 *version.Version, v2 *version.Version) bool { return v1.GreaterThan(v2) },
		">=": func(v1 *version.Version, v2 *version.Version) bool { return v1.GreaterThanOrEqual(v2) },
		"<":  func(v1 *version.Version, v2 *version.Version) bool { return v1.LessThan(v2) },
		"<=": func(v1 *version.Version, v2 *version.Version) bool { return v1.LessThanOrEqual(v2) },
	}[matcher.Comparator]
	if !ok {
		return false, fmt.Errorf("Unknown comparator: %s", matcher.Comparator)
	}

	compareTo, ok := toVersion(matcher.CompareTo)
	if !ok {
		return false, fmt.Errorf("Expected a version.  Got:\n%s", format.Object(matcher.CompareTo, 1))
	}

	actualSlice, ok := toVersions(actual)
	if !ok {
		return false, fmt.Errorf("Expected a version or a list of versions.  Got:\n%s", format.Object(actual, 1))
	}

	for _, v := range actualSlice {
		if !comparator(v, compareTo) {
			return false, nil
		}
	}

	return true, nil
}

func (matcher *BeVersionMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("to be %s", matcher.Comparator), matcher.CompareTo)
}

func (matcher *BeVersionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("not to be %s", matcher.Comparator), matcher.CompareTo)
}

func toVersion(in interface{}) (*version.Version, bool) {
	var str string
	var ok bool

	switch t := in.(type) {
	case int, float64:
		str = fmt.Sprintf("%v", t)
		ok = true
	default:
		str, ok = in.(string)
	}

	if !ok {
		return nil, false
	}

	v, err := version.NewVersion(str)
	if err != nil {
		return nil, false
	}

	return v, true
}

func toVersions(in interface{}) ([]*version.Version, bool) {
	if v, ok := toVersion(in); ok {
		return []*version.Version{v}, ok
	}

	if reflect.ValueOf(in).Kind() != reflect.Slice {
		return nil, false
	}

	out := make([]*version.Version, 0)

	if slice, ok := in.([]interface{}); ok {
		for _, ele := range slice {
			if v, ok := toVersion(ele); ok {
				out = append(out, v)
			} else {
				return nil, false
			}
		}
	} else if slice, ok := in.([]string); ok {
		for _, ele := range slice {
			if v, ok := toVersion(ele); ok {
				out = append(out, v)
			} else {
				return nil, false
			}
		}

	}

	return out, len(out) > 0
}
