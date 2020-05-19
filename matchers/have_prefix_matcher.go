package matchers

import (
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type HavePrefixMatcher struct {
	matchers.HavePrefixMatcher
}

func HavePrefix(prefix string, args ...interface{}) types.GomegaMatcher {
	return &HavePrefixMatcher{
		matchers.HavePrefixMatcher{
			Prefix: prefix,
			Args:   args,
		},
	}
}

func (matcher *HavePrefixMatcher) String() string {
	return Object(matcher.HavePrefixMatcher, 0)
}

//func (matcher *HavePrefixMatcher) String() string {
//	return fmt.Sprintf("%s{Prefix: %s}", getObjectTypeName(matcher), matcher.Prefix)
//}
//
//func getObjectTypeName(m interface{}) string {
//	return strings.Split(reflect.TypeOf(m).String(), ".")[1]
//
//}
