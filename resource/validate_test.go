package resource

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

type FakeIDer struct {
	id string
}

func (f *FakeIDer) ID() string {
	return f.id
}

var stringTests = []struct {
	in, in2 interface{}
	want    bool
}{
	{"", "", true},
	{"foo", "foo", true},
	{"foo", "bar", false},
	{"foo", "", false},
	{true, true, true},
}

func TestValidateValue(t *testing.T) {
	for _, c := range stringTests {
		inFunc := func() (interface{}, error) {
			return c.in2, nil
		}
		got := ValidateValue(&FakeIDer{""}, "", c.in, inFunc)
		if got.Result != c.want {
			t.Errorf("%+v: got %v, want %v", c, got.Result, c.want)
		}
	}
}

func TestValidateValueErr(t *testing.T) {
	for _, c := range stringTests {
		inFunc := func() (interface{}, error) {
			return c.in2, fmt.Errorf("some err")
		}
		got := ValidateValue(&FakeIDer{""}, "", c.in, inFunc)
		if got.Result != false {
			t.Errorf("%+v: got %v, want %v", c, got.Result, false)
		}
	}
}

func BenchmarkValidateValue(b *testing.B) {
	inFunc := func() (interface{}, error) {
		return "foo", nil
	}
	for n := 0; n < b.N; n++ {
		ValidateValue(&FakeIDer{""}, "", "foo", inFunc)
	}
}

var valuesTests = []struct {
	in, in2 []string
	want    bool
}{
	{[]string{""}, []string{""}, true},
	{[]string{"foo"}, []string{"foo"}, true},
	{[]string{"foo"}, []string{"bar"}, false},
	{[]string{"foo"}, []string{""}, false},
}

func TestValidateValues(t *testing.T) {
	for _, c := range valuesTests {
		inFunc := func() ([]string, error) {
			return c.in2, nil
		}
		got := ValidateValues(&FakeIDer{""}, "", c.in, inFunc)
		if got.Result != c.want {
			t.Errorf("%+v: got %v, want %v", c, got.Result, c.want)
		}
	}
}

func TestValidateValuesErr(t *testing.T) {
	for _, c := range valuesTests {
		inFunc := func() ([]string, error) {
			return c.in2, fmt.Errorf("some err")
		}
		got := ValidateValues(&FakeIDer{""}, "", c.in, inFunc)
		if got.Result != false {
			t.Errorf("%+v: got %v, want %v", c, got.Result, false)
		}
	}
}

var containsTests = []struct {
	in   []string
	in2  string
	want bool
}{
	{[]string{""}, "", true},
	{[]string{"foo"}, "foo\nbar", true},
	{[]string{"!foo"}, "foo\nbar", false},
	{[]string{"!moo"}, "foo\nbar", true},
	{[]string{"/fo.*/"}, "foo\nbar", true},
	{[]string{"!/fo.*/"}, "foo\nbar", false},
	{[]string{"!/mo.*/"}, "foo\nbar", true},
	{[]string{"foo"}, "", false},
}

func TestValidateContains(t *testing.T) {
	for _, c := range containsTests {
		inFunc := func() (io.Reader, error) {
			reader := strings.NewReader(c.in2)
			return reader, nil
		}
		got := ValidateContains(&FakeIDer{""}, "", c.in, inFunc)
		if got.Result != c.want {
			t.Errorf("%+v: got %v, want %v", c, got.Result, c.want)
		}
	}
}

func TestValidateContainsErr(t *testing.T) {
	for _, c := range containsTests {
		inFunc := func() (io.Reader, error) {
			reader := strings.NewReader(c.in2)
			return reader, fmt.Errorf("some err")
		}
		got := ValidateContains(&FakeIDer{""}, "", c.in, inFunc)
		if got.Result != false {
			t.Errorf("%+v: got %v, want %v", c, got.Result, false)
		}
	}
}
