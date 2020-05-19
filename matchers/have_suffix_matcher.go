package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type HaveSuffixMatcher struct {
	matchers.HaveSuffixMatcher
}

func HaveSuffix(prefix string, args ...interface{}) types.GomegaMatcher {
	return &HaveSuffixMatcher{
		matchers.HaveSuffixMatcher{
			Suffix: prefix,
			Args:   args,
		},
	}
}

func (matcher *HaveSuffixMatcher) String() string {
	return Object(matcher.HaveSuffixMatcher, 0)
}

//func (matcher *HaveSuffixMatcher) String() string {
//	return fmt.Sprintf("%s{Suffix: %s}", getObjectTypeName(matcher), matcher.Prefix)
//}
//
//func getObjectTypeName(m interface{}) string {
//	return strings.Split(reflect.TypeOf(m).String(), ".")[1]
//
//}
