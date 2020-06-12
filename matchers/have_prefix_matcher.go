package matchers

import (
	"encoding/json"
	"fmt"

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

func (matcher *HavePrefixMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to have prefix",
		Expected: matcher.Prefix,
	}
}

func (matcher *HavePrefixMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to have prefix",
		Expected: matcher.Prefix,
	}
}

func (matcher *HavePrefixMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["have-prefix"] = matcher.Prefix
	return json.Marshal(j)
}

func (matcher *HavePrefixMatcher) String() string {
	//return fmt.Sprintf("HavePrefix{Prefix:%s}", matcher.Prefix)
	return fmt.Sprintf("{\"have-prefix\": %q}", matcher.Prefix)
}

//func (matcher *HavePrefixMatcher) String() string {
//	return fmt.Sprintf("%s{Prefix: %s}", getObjectTypeName(matcher), matcher.Prefix)
//}
//
//func getObjectTypeName(m interface{}) string {
//	return strings.Split(reflect.TypeOf(m).String(), ".")[1]
//
//}
