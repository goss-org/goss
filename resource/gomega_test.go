package resource

import (
	"encoding/json"
	"testing"

	"github.com/aelsabbahy/goss/matchers"
	"github.com/stretchr/testify/assert"
)

var gomegaTests = []struct {
	in              string
	want            interface{}
	useNegateTester bool
}{
	// Default for simple types
	{
		in:   `"foo"`,
		want: matchers.WithSafeTransform(matchers.ToString{}, matchers.Equal("foo")),
	},
	{
		in:   `1`,
		want: matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically("==", float64(1))),
	},
	{
		in:   `true`,
		want: matchers.Equal(true),
	},
	// Default for Array
	{
		in: `["foo", "bar"]`,
		want: matchers.ContainElements(
			matchers.WithSafeTransform(matchers.ToString{}, matchers.Equal("foo")),
			matchers.WithSafeTransform(matchers.ToString{}, matchers.Equal("bar"))),
		useNegateTester: true,
	},

	// Numeric
	// Golang json escapes '>', '<' symbols, so we use 'gt', 'le' instead
	{
		in:   `{"gt": 1}`,
		want: matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically(">", float64(1))),
	},
	{
		in:   `{"ge": 1}`,
		want: matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically(">=", float64(1))),
	},
	{
		in:   `{"lt": 1}`,
		want: matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically("<", float64(1))),
	},
	{
		in:   `{"le": 1}`,
		want: matchers.WithSafeTransform(matchers.ToNumeric{}, matchers.BeNumerically("<=", float64(1))),
	},

	// String
	{
		in:   `{"have-prefix": "foo"}`,
		want: matchers.WithSafeTransform(matchers.ToString{}, matchers.HavePrefix("foo")),
	},
	{
		in:   `{"have-suffix": "foo"}`,
		want: matchers.WithSafeTransform(matchers.ToString{}, matchers.HaveSuffix("foo")),
	},
	// Regex support is based on golangs regex engine https://golang.org/pkg/regexp/syntax/
	{
		in:   `{"match-regexp": "foo"}`,
		want: matchers.WithSafeTransform(matchers.ToString{}, matchers.MatchRegexp("foo")),
	},

	// Collection
	{
		in:   `{"consist-of": ["foo"]}`,
		want: matchers.ConsistOf(matchers.WithSafeTransform(matchers.ToString{}, matchers.Equal("foo"))),
	},
	{
		in: `{"contain-element": "foo"}`,
		want: matchers.WithSafeTransform(matchers.ToArray{},
			matchers.ContainElement(
				matchers.WithSafeTransform(matchers.ToString{},
					matchers.Equal("foo")))),
	},
	{
		in:   `{"have-len": 3}`,
		want: matchers.HaveLen(3),
	},
	{
		in:   `{"have-key": "foo"}`,
		want: matchers.HaveKey(matchers.WithSafeTransform(matchers.ToString{}, matchers.Equal("foo"))),
	},

	// Negation
	{
		in:   `{"not": "foo"}`,
		want: matchers.Not(matchers.WithSafeTransform(matchers.ToString{}, matchers.Equal("foo"))),
	},
	// Complex logic
	{
		in: `{"and": ["foo", "foo"]}`,
		want: matchers.And(
			matchers.WithSafeTransform(matchers.ToString{},
				matchers.Equal("foo")),
			matchers.WithSafeTransform(matchers.ToString{},
				matchers.Equal("foo")),
		),
		useNegateTester: true,
	},
	{
		in: `{"and": [{"have-prefix": "foo"}, "foo"]}`,
		want: matchers.And(
			matchers.WithSafeTransform(matchers.ToString{},
				matchers.HavePrefix("foo")),
			matchers.WithSafeTransform(matchers.ToString{},
				matchers.Equal("foo")),
		),
		useNegateTester: true,
	},
	{
		in: `{"not": {"have-prefix": "foo"}}`,
		want: matchers.Not(
			matchers.WithSafeTransform(matchers.ToString{},
				matchers.HavePrefix("foo"))),
	},
	{
		in: `{"or": ["foo", "foo"]}`,
		want: matchers.Or(
			matchers.WithSafeTransform(matchers.ToString{},
				matchers.Equal("foo")),
			matchers.WithSafeTransform(matchers.ToString{},
				matchers.Equal("foo"))),
	},
	{
		in: `{"not": {"and": [{"have-prefix": "foo"}]}}`,
		want: matchers.Not(
			matchers.And(
				matchers.WithSafeTransform(matchers.ToString{},
					matchers.HavePrefix("foo")))),
	},

	// Semver Constraint
	{
		in:   `{"semver-constraint": "> 1.0.0"}`,
		want: matchers.BeSemverConstraint("> 1.0.0"),
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
		gomegaTestEqual(t, got, c.want, c.useNegateTester, c.in)
	}
}

func gomegaTestEqual(t *testing.T, got, want interface{}, useNegateTester bool, in string) {
	assert.Equal(t, got, want)
}
