package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
)

type ConsistOfMatcher struct {
	matchers.ConsistOfMatcher
}

func ConsistOf(elements ...interface{}) GossMatcher {
	return &ConsistOfMatcher{
		matchers.ConsistOfMatcher{
			Elements: elements,
		},
	}
}

func (matcher *ConsistOfMatcher) FailureResult(actual interface{}) MatcherResult {
	missingElements := getUnexported(matcher, "missingElements")
	extraElements := getUnexported(matcher, "extraElements")
	return MatcherResult{
		Actual:          actual,
		Message:         "to consist of",
		Expected:        matcher.Elements,
		MissingElements: missingElements,
		ExtraElements:   extraElements,
	}
}

func (matcher *ConsistOfMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to consist of",
		Expected: matcher.Elements,
	}
}

func (matcher *ConsistOfMatcher) String() string {
	return Object(matcher.ConsistOfMatcher, 0)
}
func (matcher *ConsistOfMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["consist-of"] = matcher.Elements
	return json.Marshal(j)
}
