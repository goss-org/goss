package resource

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

type FakeResource struct {
	id string
}

func (f *FakeResource) ID() string {
	return f.id
}
func (f *FakeResource) GetTitle() string { return "title" }

//func (f *FakeResource) GetMeta() meta    { return map[string]interface{}{"foo": "bar"} }
func (f *FakeResource) GetMeta() meta { return meta{"foo": "bar"} }

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
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc)
		if got.Successful != c.want {
			t.Errorf("%+v: got %v, want %v", c, got.Successful, c.want)
		}
	}
}

func TestValidateValueErr(t *testing.T) {
	for _, c := range stringTests {
		inFunc := func() (interface{}, error) {
			return c.in2, fmt.Errorf("some err")
		}
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc)
		if got.Successful != false {
			t.Errorf("%+v: got %v, want %v", c, got.Successful, false)
		}
	}
}

func BenchmarkValidateValue(b *testing.B) {
	inFunc := func() (interface{}, error) {
		return "foo", nil
	}
	for n := 0; n < b.N; n++ {
		ValidateValue(&FakeResource{""}, "", "foo", inFunc)
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
	{[]string{`/\s/tmp\b/`}, "test /tmp bar", true},
}

func TestValidateContains(t *testing.T) {
	for _, c := range containsTests {
		inFunc := func() (io.Reader, error) {
			reader := strings.NewReader(c.in2)
			return reader, nil
		}
		got := ValidateContains(&FakeResource{""}, "", c.in, inFunc)
		if got.Successful != c.want {
			t.Errorf("%+v: got %v, want %v", c, got.Successful, c.want)
		}
	}
}

func TestValidateContainsErr(t *testing.T) {
	for _, c := range containsTests {
		inFunc := func() (io.Reader, error) {
			reader := strings.NewReader(c.in2)
			return reader, fmt.Errorf("some err")
		}
		got := ValidateContains(&FakeResource{""}, "", c.in, inFunc)
		if got.Successful != false {
			t.Errorf("%+v: got %v, want %v", c, got.Successful, false)
		}
	}
}

func TestValidateContainsBadRegexErr(t *testing.T) {
	inFunc := func() (io.Reader, error) {
		reader := strings.NewReader("dummy")
		return reader, nil
	}
	got := ValidateContains(&FakeResource{""}, "", []string{"/*\\.* @@.*/"}, inFunc)
	if got.Err == nil {
		t.Errorf("Expected bad regex to raise error, got nil")
	}
}
