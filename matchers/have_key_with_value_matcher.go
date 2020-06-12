package matchers

import (
	"encoding/json"

	"github.com/icza/dyno"
	"github.com/onsi/gomega/matchers"
)

type HaveKeyWithValueMatcher struct {
	matchers.HaveKeyWithValueMatcher
}

func HaveKeyWithValue(key interface{}, value interface{}) GossMatcher {
	return &HaveKeyWithValueMatcher{
		matchers.HaveKeyWithValueMatcher{
			Key:   key,
			Value: value,
		},
	}
}

func (matcher *HaveKeyWithValueMatcher) FailureResult(actual interface{}) MatcherResult {
	expect := make(map[interface{}]interface{}, 1)
	expect[matcher.Key] = matcher.Value
	return MatcherResult{
		Actual:   actual,
		Message:  "to have {key: value} matching",
		Expected: expect,
	}
}

func (matcher *HaveKeyWithValueMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	expect := make(map[interface{}]interface{}, 1)
	expect[matcher.Key] = matcher.Value
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have {key: value} matching",
		Expected: matcher.Key,
	}
}

func (matcher *HaveKeyWithValueMatcher) MarshalJSON() ([]byte, error) {
	expect := make(map[interface{}]interface{}, 1)
	expect[matcher.Key] = matcher.Value
	j := make(map[string]interface{})
	j["have-key-with-value"] = expect
	return json.Marshal(dyno.ConvertMapI2MapS(j))
}

func (matcher *HaveKeyWithValueMatcher) String() string {
	return ""
	return Object(matcher.HaveKeyWithValueMatcher, 0)
}
