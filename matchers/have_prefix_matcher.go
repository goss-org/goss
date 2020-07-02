package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type HavePrefixMatcher struct {
	matchers.HavePrefixMatcher
}

func HavePrefix(prefix string, args ...interface{}) GossMatcher {
	return &HavePrefixMatcher{
		matchers.HavePrefixMatcher{
			Prefix: prefix,
			Args:   args,
		},
	}
}

func (m *HavePrefixMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have prefix",
		Expected: m.Prefix,
	}
}

func (m *HavePrefixMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have prefix",
		Expected: m.Prefix,
	}
}

func (m *HavePrefixMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["have-prefix"] = m.Prefix
	return json.Marshal(j)
}
