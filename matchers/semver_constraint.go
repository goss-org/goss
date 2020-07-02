package matchers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/blang/semver"
	"github.com/onsi/gomega/format"
)

func BeSemverConstraint(constraint interface{}) GossMatcher {
	return &BeSemverConstraintMatcher{
		Constraint: constraint,
	}
}

type BeSemverConstraintMatcher struct {
	Constraint interface{}
}

func (m *BeSemverConstraintMatcher) Match(actual interface{}) (success bool, err error) {
	constraint, ok := toConstraint(m.Constraint)
	if !ok {
		return false, fmt.Errorf("Expected a valid semver constraint.  Got:\n%s", format.Object(m.Constraint, 1))
	}

	actualSlice, ok := toVersions(actual)
	if !ok {
		return false, fmt.Errorf("Expected a single or list of semver valid version(s).  Got:\n%s", format.Object(actual, 1))
	}

	for _, v := range actualSlice {
		if !constraint(*v) {
			return false, nil
		}
	}

	return true, nil
}

func (m *BeSemverConstraintMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to satisfy constraint",
		Expected: m.Constraint,
	}
}

func (m *BeSemverConstraintMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to satisfy constraint",
		Expected: m.Constraint,
	}
}

func (m *BeSemverConstraintMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("to be %s", m.Constraint))
}

func (m *BeSemverConstraintMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("not to be %s", m.Constraint))
}

//func (m *BeSemverConstraintMatcher) String() string {
//	return Object(m, 0)
//}

func toConstraint(in interface{}) (semver.Range, bool) {
	str, ok := in.(string)
	if !ok {
		return nil, false
	}

	out, err := semver.ParseRange(str)
	return out, err == nil
}

func toVersion(in interface{}) (*semver.Version, bool) {
	str, ok := in.(string)
	if !ok {
		return nil, false
	}

	v, err := semver.Parse(str)
	if err != nil {
		return nil, false
	}

	return &v, true
}

func toVersions(in interface{}) ([]*semver.Version, bool) {
	if v, ok := toVersion(in); ok {
		return []*semver.Version{v}, ok
	}

	if reflect.ValueOf(in).Kind() != reflect.Slice {
		return nil, false
	}

	out := make([]*semver.Version, 0)

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

func (m *BeSemverConstraintMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["semver-constraint"] = m.Constraint
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(j)
	if err != nil {
		return nil, nil
	}
	b := buffer.Bytes()
	return b, nil
}

func (m *BeSemverConstraintMatcher) String() string {
	return format.Object(m, 0)
}
