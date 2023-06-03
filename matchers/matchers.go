package matchers

import (
	"reflect"
	"unsafe"

	"github.com/onsi/gomega/types"
)

type GossMatcher interface {
	// This is needed due to oMegaMatcher test in some of the GomegaMatcher logic
	types.GomegaMatcher
	//Match(actual interface{}) (success bool, err error)
	FailureResult(actual interface{}) MatcherResult
	NegatedFailureResult(actual interface{}) MatcherResult
	//json.Marshaler
}

type MatcherResult struct {
	Actual             interface{}
	Message            string
	Expected           interface{}
	MissingElements    interface{}
	ExtraElements      interface{}
	TransformerChain   []Transformer
	UntransformedValue interface{}
}

func getUnexported(i interface{}, field string) interface{} {
	rs := reflect.ValueOf(i).Elem()
	rf := rs.FieldByName(field)
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	return rf.Interface()
}

type fakeOmegaMatcher struct{}

// FailureMessage is a stub to honor omegaMatcher interface
func (m *fakeOmegaMatcher) FailureMessage(_ interface{}) (message string) {
	return ""
}

// NegatedFailureMessage is a stub to honor omegaMatcher interface
func (m *fakeOmegaMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return ""
}
