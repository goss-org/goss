package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var gomegaTests = []struct {
	in              string
	want            interface{}
	useNegateTester bool
}{
	// Default for simple types
	{
		in:   `"foo"`,
		want: gomega.Equal("foo"),
	},
	{
		in:   `1`,
		want: gomega.Equal(float64(1)),
	},
	{
		in:   `true`,
		want: gomega.Equal(true),
	},
	// Default for Array
	{
		in:              `["foo", "bar"]`,
		want:            gomega.And(gomega.ContainElement("foo"), gomega.ContainElement("bar")),
		useNegateTester: true,
	},

	// Numeric
	// Golang json escapes '>', '<' symbols, so we use 'gt', 'le' instead
	{
		in:   `{"gt": 1}`,
		want: gomega.BeNumerically(">", float64(1)),
	},
	{
		in:   `{"ge": 1}`,
		want: gomega.BeNumerically(">=", float64(1)),
	},
	{
		in:   `{"lt": 1}`,
		want: gomega.BeNumerically("<", float64(1)),
	},
	{
		in:   `{"le": 1}`,
		want: gomega.BeNumerically("<=", float64(1)),
	},

	// String
	{
		in:   `{"have-prefix": "foo"}`,
		want: gomega.HavePrefix("foo"),
	},
	{
		in:   `{"have-suffix": "foo"}`,
		want: gomega.HaveSuffix("foo"),
	},
	{
		in:   `{"match-regexp": "foo"}`,
		want: gomega.MatchRegexp("foo"),
	},

	// Collection
	{
		in:   `{"consist-of": ["foo"]}`,
		want: gomega.ConsistOf(gomega.Equal("foo")),
	},
	{
		in:   `{"contain-element": "foo"}`,
		want: gomega.ContainElement("foo"),
	},
	{
		in:   `{"have-len": 3}`,
		want: gomega.HaveLen(3),
	},

	// Negation
	{
		in:   `{"not": "foo"}`,
		want: gomega.Not(gomega.Equal("foo")),
	},
	// Complex logic
	{
		in:              `{"and": ["foo", "foo"]}`,
		want:            gomega.And(gomega.Equal("foo"), gomega.Equal("foo")),
		useNegateTester: true,
	},
	{
		in:              `{"and": [{"have-prefix": "foo"}, "foo"]}`,
		want:            gomega.And(gomega.HavePrefix("foo"), gomega.Equal("foo")),
		useNegateTester: true,
	},
	{
		in:   `{"not": {"have-prefix": "foo"}}`,
		want: gomega.Not(gomega.HavePrefix("foo")),
	},
	{
		in:   `{"or": ["foo", "foo"]}`,
		want: gomega.Or(gomega.Equal("foo"), gomega.Equal("foo")),
	},
	{
		in:   `{"not": {"and": [{"have-prefix": "foo"}]}}`,
		want: gomega.Not(gomega.And(gomega.HavePrefix("foo"))),
	},
}

func TestMatcherToGomegaMatcher(t *testing.T) {
	for _, c := range gomegaTests {
		var dat interface{}
		if err := json.Unmarshal([]byte(c.in), &dat); err != nil {
			t.Fatal(err)
		}
		got, err := matcherToGomegaMatcher(dat)
		if err != nil {
			t.Fatal(err)
		}
		gomegaTestEqual(t, got, c.want, c.useNegateTester)
	}
}

func gomegaTestEqual(t *testing.T, got, want interface{}, useNegateTester bool) {
	if !gomegaEqual(got, want, useNegateTester) {
		t.Errorf("got %T %v, want %T %v", got, got, want, want)
	}
}
func gomegaEqual(g, w interface{}, negateTester bool) bool {
	gotT := reflect.TypeOf(g)
	wantT := reflect.TypeOf(w)
	got := g.(types.GomegaMatcher)
	want := w.(types.GomegaMatcher)
	var gotMessage string
	var wantMessage string
	if negateTester {
		gotMessage = got.NegatedFailureMessage("foo")
		wantMessage = want.NegatedFailureMessage("foo")
	} else {
		gotMessage = got.FailureMessage("foo")
		wantMessage = want.FailureMessage("foo")
	}
	fmt.Println("got:", gotMessage)
	fmt.Println("want:", wantMessage)

	return gotT == wantT &&
		gotMessage == wantMessage
}
