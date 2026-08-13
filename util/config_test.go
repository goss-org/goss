package util

import (
	"reflect"
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

func TestWithVarsFiles(t *testing.T) {
	files := []string{"/nonexisting"}
	c, err := NewConfig(WithVarsFiles(files))
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(c.VarsFiles, files) {
		t.Fatalf("expected %s got %q", files, c.VarsFiles)
	}

	files = []string{"/nonexisting", "/second", "third"}
	c, err = NewConfig(WithVarsFiles(files))
	if err != nil {
		t.Fatal(err.Error())
	}

	if !reflect.DeepEqual(c.VarsFiles, files) {
		t.Fatalf("expected %s got %q", files, c.VarsFiles)
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
