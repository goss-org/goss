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
