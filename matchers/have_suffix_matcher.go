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
	j["have-suffix"] = m.Suffix
	return json.Marshal(j)
}
