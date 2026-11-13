package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type HavePrefixMatcher struct {
	matchers.HavePrefixMatcher
}

func HavePrefix(prefix string, args ...any) GossMatcher {
	return &HavePrefixMatcher{
		matchers.HavePrefixMatcher{
			Prefix: prefix,
			Args:   args,
		},
	}
}

func (m *HavePrefixMatcher) FailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have prefix",
		Expected: m.Prefix,
	}
}

func (m *HavePrefixMatcher) NegatedFailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have prefix",
		Expected: m.Prefix,
	}
}

func (m *HavePrefixMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]any)
	j["have-prefix"] = m.Prefix
	return json.Marshal(j)
}
