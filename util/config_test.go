package util

import (
	"testing"
)

func TestWithVarsBytes(t *testing.T) {
	vs := `{"hello":"world"}`
	c, err := NewConfig(WithVarsBytes([]byte(vs)))
	if err != nil {
		t.Fatal(err.Error())
	}

	if c.VarsInline != vs {
		t.Fatalf("expected %q got %q", vs, c.VarsInline)
	}
}

func TestWithVarsString(t *testing.T) {
	vs := `{"hello":"world"}`
	c, err := NewConfig(WithVarsString(vs))
	if err != nil {
		t.Fatal(err.Error())
	}

	if c.VarsInline != vs {
		t.Fatalf("expected %q got %q", vs, c.VarsInline)
	}
}

func TestWithVarsFile(t *testing.T) {
	c, err := NewConfig(WithVarsFile("/nonexisting"))
	if err != nil {
		t.Fatal(err.Error())
	}

	if c.Vars != "/nonexisting" {
		t.Fatalf("expected '/nonexisting' got %q", c.Vars)
	}
}

func TestWithVarsData(t *testing.T) {
	c, err := NewConfig(WithVarsData(map[string]string{"hello": "world"}))
	if err != nil {
		t.Fatal(err.Error())
	}

	if c.VarsInline != `{"hello":"world"}` {
		t.Fatalf("expected %q got %q", `{"hello":"world"}`, c.VarsInline)
	}
}

func TestWithIncludeMarks(t *testing.T) {
	c, err := NewConfig(WithIncludeMarks("critical", "network"))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.IncludeMarks) != 2 || c.IncludeMarks[0] != "critical" || c.IncludeMarks[1] != "network" {
		t.Fatalf("unexpected IncludeMarks: %v", c.IncludeMarks)
	}
}

func TestWithExcludeMarks(t *testing.T) {
	c, err := NewConfig(WithExcludeMarks("slow", "flaky"))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.ExcludeMarks) != 2 || c.ExcludeMarks[0] != "slow" || c.ExcludeMarks[1] != "flaky" {
		t.Fatalf("unexpected ExcludeMarks: %v", c.ExcludeMarks)
	}
}

func TestWithMarksMultipleCallsAppend(t *testing.T) {
	c, err := NewConfig(WithIncludeMarks("a"), WithIncludeMarks("b", "c"))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.IncludeMarks) != 3 {
		t.Fatalf("expected 3 marks got %v", c.IncludeMarks)
	}
}

func TestParseMarksParam(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{",,,", nil},
		{"critical", []string{"critical"}},
		{"critical,network", []string{"critical", "network"}},
		{"  critical , network  ", []string{"critical", "network"}},
		{"critical, ,network", []string{"critical", "network"}},
		{"a,b,c,d", []string{"a", "b", "c", "d"}},
	}
	for _, tc := range cases {
		got := ParseMarksParam(tc.in)
		if len(got) != len(tc.want) {
			t.Errorf("ParseMarksParam(%q) length = %d, want %d (got=%v)", tc.in, len(got), len(tc.want), got)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("ParseMarksParam(%q)[%d] = %q, want %q", tc.in, i, got[i], tc.want[i])
			}
		}
	}
}
