package matchers

import (
	"fmt"

	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
	"github.com/sanity-io/litter"
)

type EqualMatcher struct {
	matchers.EqualMatcher
}

func Equal(element interface{}) types.GomegaMatcher {
	return &EqualMatcher{
		matchers.EqualMatcher{
			Expected: element,
		},
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
