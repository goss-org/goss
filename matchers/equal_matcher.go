package matchers

import (
	"encoding/json"
	"fmt"

	"github.com/onsi/gomega/matchers"
	"github.com/sanity-io/litter"
)

type EqualMatcher struct {
	matchers.EqualMatcher
}

func Equal(element interface{}) GossMatcher {
	return &EqualMatcher{
		matchers.EqualMatcher{
			Expected: element,
		},
	}
}

func (matcher *EqualMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to equal",
		Expected: matcher.Expected,
	}
}

func (matcher *EqualMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to equal",
		Expected: matcher.Expected,
	}
}

func (matcher *EqualMatcher) GoString() string {
	sq := litter.Options{Compact: true, StripPackageNames: true}
	return sq.Sdump(matcher.EqualMatcher)
}
func (matcher *EqualMatcher) String() string {
	//sq := litter.Options{Compact: true, StripPackageNames: true}
	//return sq.Sdump(matcher.EqualMatcher)
	//return fmt.Sprintf("EqualMatcher{Expected:\"%s\"}", matcher.Expected)
	return fmt.Sprintf("%q", matcher.Expected)
}

func (matcher *EqualMatcher) MarshalJSON() ([]byte, error) {
	return json.Marshal(matcher.Expected)
}
