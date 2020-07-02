package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type HaveSuffixMatcher struct {
	matchers.HaveSuffixMatcher
}

func HaveSuffix(prefix string, args ...interface{}) GossMatcher {
	return &HaveSuffixMatcher{
		matchers.HaveSuffixMatcher{
			Suffix: prefix,
			Args:   args,
		},
	}
}

func (m *HaveSuffixMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have suffix",
		Expected: m.Suffix,
	}
}

func (m *HaveSuffixMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have suffix",
		Expected: m.Suffix,
	}
}

func (m *HaveSuffixMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["have-prefix"] = m.Suffix
	return json.Marshal(j)
}

func (m *HaveSuffixMatcher) String() string {
	return Object(m.HaveSuffixMatcher, 0)
}

//func (m *HaveSuffixMatcher) String() string {
//	return fmt.Sprintf("%s{Suffix: %s}", getObjectTypeName(m), m.Prefix)
//}
//
//func getObjectTypeName(m interface{}) string {
//	return strings.Split(reflect.TypeOf(m).String(), ".")[1]
//
//}
