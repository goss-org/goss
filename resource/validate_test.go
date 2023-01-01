package resource

import (
	"encoding/json"
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

func (f *FakeResource) GetMeta() meta { return meta{"foo": "bar"} }

var stringTests = []struct {
	in, in2 any
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
		inFunc := func() (any, error) {
			return c.in2, nil
		}
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc, false)
		if got.Successful != c.want {
			t.Errorf("%+v: got %v, want %v", c, got.Successful, c.want)
		}
	}
}

func TestValidateValueErr(t *testing.T) {
	for _, c := range stringTests {
		inFunc := func() (any, error) {
			return c.in2, fmt.Errorf("some err")
		}
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc, false)
		if got.Successful != false {
			t.Errorf("%+v: got %v, want %v", c, got.Successful, false)
		}
	}
}

func TestValidateValueSkip(t *testing.T) {
	for _, c := range stringTests {
		inFunc := func() (any, error) {
			return c.in2, nil
		}
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc, true)
		if got.Result != SKIP {
			t.Errorf("%+v: got %v, want %v", c, got.Result, SKIP)
		}
	}
}

func BenchmarkValidateValue(b *testing.B) {
	inFunc := func() (any, error) {
		return "foo", nil
	}
	for n := 0; n < b.N; n++ {
		ValidateValue(&FakeResource{""}, "", "foo", inFunc, false)
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
		got := ValidateContains(&FakeResource{""}, "", c.in, inFunc, false)
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
		got := ValidateContains(&FakeResource{""}, "", c.in, inFunc, false)
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
	got := ValidateContains(&FakeResource{""}, "", []string{"/*\\.* @@.*/"}, inFunc, false)
	if got.Err == nil {
		t.Errorf("Expected bad regex to raise error, got nil")
	}
}

func TestValidateContainsSkip(t *testing.T) {
	for _, c := range containsTests {
		inFunc := func() (io.Reader, error) {
			reader := strings.NewReader(c.in2)
			return reader, nil
		}
		got := ValidateContains(&FakeResource{""}, "", c.in, inFunc, true)
		if got.Result != SKIP {
			t.Errorf("%+v: got %v, want %v", c, got.Result, SKIP)
		}
	}
}

func TestResultMarshaling(t *testing.T) {
	inFunc := func() (io.Reader, error) {
		return nil, fmt.Errorf("dummy error")
	}
	res := ValidateContains(&FakeResource{}, "", []string{"x"}, inFunc, false)
	if res.Err == nil {
		t.Fatalf("Expected to receive an error")
	}
	if res.Err.Error() != "dummy error" {
		t.Fatalf("expected to receive 'dummy error', got: %v", res.Err.Error())
	}

	rj, _ := json.Marshal(res)
	res = TestResult{}
	err := json.Unmarshal(rj, &res)
	if err != nil {
		t.Fatalf("could not unmarshal result: %v", err)
	}

	if res.Err == nil {
		t.Fatalf("Expected to receive an error")
	}
	if res.Err.Error() != "dummy error" {
		t.Fatalf("expected to receive 'dummy error', got: %v", res.Err.Error())
	}
}
