package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type HaveLenMatcher struct {
	matchers.HaveLenMatcher
}

func HaveLen(count int) types.GomegaMatcher {
	return &HaveLenMatcher{
		matchers.HaveLenMatcher{
			Count: count,
		},
	}
}

func (matcher *HaveLenMatcher) String() string {
	return Object(matcher.HaveLenMatcher, 0)
}

//func (matcher *HaveLenMatcher) String() string {
//	n := fmt.Sprintf("%#v", matcher.HaveLenMatcher)
//	ss := strings.Split(n, ".")
//	s := ss[len(ss)-1]
//	return s
//}
