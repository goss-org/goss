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
	want    int
}{
	{"", "", SUCCESS},
	{"foo", "foo", SUCCESS},
	{"foo", "bar", FAIL},
	{"foo", "", FAIL},
	{true, true, SUCCESS},
}

func TestValidateValue(t *testing.T) {
	for _, c := range stringTests {
		inFunc := func() (any, error) {
			return c.in2, nil
		}
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc, false)
		if got.Result != c.want {
			t.Errorf("%+v: got %v, want %v", c, got.Result, c.want)
		}
	}
}

func TestValidateValueErr(t *testing.T) {
	for _, c := range stringTests {
		inFunc := func() (any, error) {
			return c.in2, fmt.Errorf("some err")
		}
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc, false)
		if got.Result != FAIL {
			t.Errorf("%+v: got %v, want %v", c, got.Result, FAIL)
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
	in   []interface{}
	in2  string
	want int
}{
	{[]interface{}{""}, "", SUCCESS},
	{[]interface{}{"foo"}, "foo\nbar", SUCCESS},
	{[]interface{}{"!foo"}, "foo\nbar", FAIL},
	{[]interface{}{"!moo"}, "foo\nbar", SUCCESS},
	{[]interface{}{"/fo.*/"}, "foo\nbar", SUCCESS},
	{[]interface{}{"!/fo.*/"}, "foo\nbar", FAIL},
	{[]interface{}{"!/mo.*/"}, "foo\nbar", SUCCESS},
	{[]interface{}{"foo"}, "", FAIL},
	{[]interface{}{`/\s/tmp\b/`}, "test /tmp bar", SUCCESS},
}

func TestValidateContains(t *testing.T) {
	for _, c := range containsTests {
		inFunc := func() (io.Reader, error) {
			reader := strings.NewReader(c.in2)
			return reader, nil
		}
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc, false)
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
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc, false)
		if got.Result != FAIL {
			t.Errorf("%+v: got %v, want %v", c, got.Result, FAIL)
		}
	}
}

func TestValidateContainsBadRegexErr(t *testing.T) {
	inFunc := func() (io.Reader, error) {
		reader := strings.NewReader("dummy")
		return reader, nil
	}
	got := ValidateValue(&FakeResource{""}, "", []interface{}{"/*\\.* @@.*/"}, inFunc, false)
	if got.Err == nil {
		t.Errorf("Expected bad regex to raise error, got nil")
	}
}

// TestValidateContainsFailureActual asserts that when a file-contents check fails,
// the MatcherResult.Actual field contains the readable file content — not the
// Go internal type representation "object: *bytes.Reader" / "object: *os.File".
func TestValidateContainsFailureActual(t *testing.T) {
	fileContent := "Banner /etc/issue.net\nLogLevel INFO\n"
	missingPattern := "nonexistent-pattern-xyz"

	inFunc := func() (io.Reader, error) {
		return strings.NewReader(fileContent), nil
	}
	got := ValidateValue(&FakeResource{""}, "contents", []interface{}{missingPattern}, inFunc, false)

	if got.Result != FAIL {
		t.Fatalf("expected FAIL, got %v", got.Result)
	}
	actual, ok := got.MatcherResult.Actual.(string)
	if !ok {
		t.Fatalf("MatcherResult.Actual must be a string, got %T: %v", got.MatcherResult.Actual, got.MatcherResult.Actual)
	}
	if strings.Contains(actual, "object:") {
		t.Errorf("MatcherResult.Actual must not contain Go type repr, got: %q", actual)
	}
	if !strings.Contains(actual, "Banner") {
		t.Errorf("MatcherResult.Actual must contain file content, got: %q", actual)
	}
}

func TestValidateContainsSkip(t *testing.T) {
	for _, c := range containsTests {
		inFunc := func() (io.Reader, error) {
			reader := strings.NewReader(c.in2)
			return reader, nil
		}
		got := ValidateValue(&FakeResource{""}, "", c.in, inFunc, true)
		if got.Result != SKIP {
			t.Errorf("%+v: got %v, want %v", c, got.Result, SKIP)
		}
	}
}

func TestResultMarshaling(t *testing.T) {
	inFunc := func() (io.Reader, error) {
		return nil, fmt.Errorf("dummy error")
	}
	res := ValidateValue(&FakeResource{}, "", []string{"x"}, inFunc, false)
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
