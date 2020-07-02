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

func (m *EqualMatcher) FailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "to equal",
		Expected: m.Expected,
	}
}

func (m *EqualMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to equal",
		Expected: m.Expected,
	}
}

func (m *EqualMatcher) GoString() string {
	sq := litter.Options{Compact: true, StripPackageNames: true}
	return sq.Sdump(m.EqualMatcher)
}
func (m *EqualMatcher) String() string {
	//sq := litter.Options{Compact: true, StripPackageNames: true}
	//return sq.Sdump(m.EqualMatcher)
	//return fmt.Sprintf("EqualMatcher{Expected:\"%s\"}", m.Expected)
	return fmt.Sprintf("%q", m.Expected)
}

func (m *EqualMatcher) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Expected)
}
