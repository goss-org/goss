package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/matchers"
)

type ContainSubstringMatcher struct {
	matchers.ContainSubstringMatcher
}

func ContainSubstring(substr string, args ...interface{}) GossMatcher {
	return &ContainSubstringMatcher{
		matchers.ContainSubstringMatcher{
			Substr: substr,
			Args:   args,
		},
	}
}

func (matcher *ContainSubstringMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to contain substring",
		Expected: matcher.Substr,
	}
}

func (matcher *ContainSubstringMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to contain substring",
		Expected: matcher.Substr,
	}
}

func (matcher *ContainSubstringMatcher) String() string {
	return format.Object(matcher.ContainSubstringMatcher, 0)
}

func (matcher *ContainSubstringMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["contain-substring"] = matcher.Substr
	return json.Marshal(j)
}
