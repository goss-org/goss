package matchers

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/matchers/support/goraph/bipartitegraph"
)

type GomegaContainElementsMatcher struct {
	Elements        []interface{}
	MissingElements []interface{}
}

func (matcher *GomegaContainElementsMatcher) Match(actual interface{}) (success bool, err error) {
	if !isArrayOrSlice(actual) && !isMap(actual) {
		return false, fmt.Errorf("ContainElements matcher expects an array/slice/map.  Got:\n%s", format.Object(actual, 1))
	}

	matchers := gmatchers(matcher.Elements)
	bipartiteGraph, err := bipartitegraph.NewBipartiteGraph(valuesOf(actual), matchers, neighbours)
	if err != nil {
		return false, err
	}

	edges := bipartiteGraph.LargestMatching()
	if len(edges) == len(matchers) {
		return true, nil
	}

	_, missingMatchers := bipartiteGraph.FreeLeftRight(edges)
	matcher.MissingElements = equalMatchersToElements(missingMatchers)

	return false, nil
}

func (matcher *GomegaContainElementsMatcher) FailureMessage(actual interface{}) (message string) {
	message = format.Message(actual, "to contain elements", matcher.Elements)
	return appendMissingElements(message, matcher.MissingElements)
}

func (matcher *GomegaContainElementsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to contain elements", matcher.Elements)
}
