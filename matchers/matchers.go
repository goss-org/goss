package matchers

import (
	"encoding/json"
	"reflect"
	"unsafe"

	"github.com/onsi/gomega/types"
)

type GossMatcher interface {
	// This is needed due to oMegaMatcher test in some of the GomegaMatcher logic
	types.GomegaMatcher
	//Match(actual interface{}) (success bool, err error)
	FailureResult(actual any) MatcherResult
	NegatedFailureResult(actual any) MatcherResult
	// This doesn't seem to make a difference, maybe not needed
	json.Marshaler
}

type MatcherResult struct {
	Actual             any           `json:"actual"`
	Message            string        `json:"message"`
	Expected           any           `json:"expected"`
	MissingElements    any           `json:"missing-elements"`
	FoundElements      any           `json:"found-elements"`
	ExtraElements      any           `json:"extra-elements"`
	TransformerChain   []Transformer `json:"transform-chain"`
	UntransformedValue any           `json:"untransformed-value"`
}

func getUnexported(i any, field string) any {
	rs := reflect.ValueOf(i).Elem()
	rf := rs.FieldByName(field)
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	return rf.Interface()
}

type fakeOmegaMatcher struct{}

// FailureMessage is a stub to honor omegaMatcher interface
func (m *fakeOmegaMatcher) FailureMessage(_ any) (message string) {
	return ""
}

// NegatedFailureMessage is a stub to honor omegaMatcher interface
func (m *fakeOmegaMatcher) NegatedFailureMessage(_ any) (message string) {
	return ""
}
