package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/goss-org/goss/matchers"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var gomegaTests = []struct {
	in              string
	want            any
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
	// Regex support is based on golangs regex engine https://golang.org/pkg/regexp/syntax/
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
		want: gomega.ContainElement(gomega.Equal("foo")),
	},
	{
		in:   `{"have-len": 3}`,
		want: gomega.HaveLen(3),
	},
	{
		in: `{"have-key-with-value": { "foo": 1, "bar": "baz" }}`,
		// Keys are sorted and then passed to gomega.And so the order
		// of the conditions in this `want` is important
		want: gomega.And(
			gomega.HaveKeyWithValue("bar", gomega.Equal("baz")),
			gomega.HaveKeyWithValue("foo", gomega.Equal(1)),
		),
		useNegateTester: true,
	},
	{
		in:   `{"have-key": "foo"}`,
		want: gomega.HaveKey(gomega.Equal("foo")),
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

	// Semver Constraint
	{
		in:   `{"semver-constraint": "> 1.0.0"}`,
		want: matchers.BeSemverConstraint("> 1.0.0"),
	},
}

func TestMatcherToGomegaMatcher(t *testing.T) {
	for _, c := range gomegaTests {
		var dat any
		if err := json.Unmarshal([]byte(c.in), &dat); err != nil {
			t.Fatal(err)
		}
		got, err := matcherToGomegaMatcher(dat)
		if err != nil {
			t.Fatal(err)
		}
		gomegaTestEqual(t, got, c.want, c.useNegateTester, c.in)
	}
}

func gomegaTestEqual(t *testing.T, got, want any, useNegateTester bool, in string) {
	if !gomegaEqual(got, want, useNegateTester) {
		t.Errorf("For input '%s': got %T %v, want %T %v", in, got, got, want, want)
	}
}
func gomegaEqual(g, w any, negateTester bool) bool {
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
	gotMessage = sanitizeMatcherText(gotMessage)
	wantMessage = sanitizeMatcherText(wantMessage)
	fmt.Println("got:", gotMessage)
	fmt.Println("want:", wantMessage)

	return gotT == wantT &&
		gotMessage == wantMessage
}

func sanitizeMatcherText(s string) string {
	r := regexp.MustCompile("[0-9]x[a-z0-9]{10}")
	return r.ReplaceAllString(s, "")
}
