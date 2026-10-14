package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type MatchRegexpMatcher struct {
	matchers.MatchRegexpMatcher
}

func MatchRegexp(regexp string, args ...any) GossMatcher {
	return &MatchRegexpMatcher{
		matchers.MatchRegexpMatcher{
			Regexp: regexp,
			Args:   args,
		},
	}
}

func (m *MatchRegexpMatcher) FailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to match regular expression",
		Expected: m.Regexp,
	}
}

func (m *MatchRegexpMatcher) NegatedFailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to match regular expression",
		Expected: m.Regexp,
	}
}

func (m *MatchRegexpMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]any)
	j["match-regexp"] = m.Regexp
	return json.Marshal(j)
}
