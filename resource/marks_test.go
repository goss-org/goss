package resource

import (
	"encoding/json"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

// TestResourceMarksRoundTrip ensures the Marks field round-trips through both
// JSON and YAML encoders for every resource type that supports it.
func TestResourceMarksRoundTrip(t *testing.T) {
	t.Parallel()

	marks := []string{"critical", "network"}

	cases := []struct {
		name   string
		newRes func() interface {
			GetMarks() []string
		}
	}{
		{"Addr", func() interface{ GetMarks() []string } { return &Addr{Marks: marks} }},
		{"Command", func() interface{ GetMarks() []string } { return &Command{Marks: marks} }},
		{"DNS", func() interface{ GetMarks() []string } { return &DNS{Marks: marks} }},
		{"File", func() interface{ GetMarks() []string } { return &File{Marks: marks} }},
		{"Gossfile", func() interface{ GetMarks() []string } { return &Gossfile{Marks: marks} }},
		{"Group", func() interface{ GetMarks() []string } { return &Group{Marks: marks} }},
		{"HTTP", func() interface{ GetMarks() []string } { return &HTTP{Marks: marks} }},
		{"Interface", func() interface{ GetMarks() []string } { return &Interface{Marks: marks} }},
		{"KernelParam", func() interface{ GetMarks() []string } { return &KernelParam{Marks: marks} }},
		{"Matching", func() interface{ GetMarks() []string } { return &Matching{Marks: marks} }},
		{"Mount", func() interface{ GetMarks() []string } { return &Mount{Marks: marks} }},
		{"Package", func() interface{ GetMarks() []string } { return &Package{Marks: marks} }},
		{"Port", func() interface{ GetMarks() []string } { return &Port{Marks: marks} }},
		{"Process", func() interface{ GetMarks() []string } { return &Process{Marks: marks} }},
		{"Service", func() interface{ GetMarks() []string } { return &Service{Marks: marks} }},
		{"User", func() interface{ GetMarks() []string } { return &User{Marks: marks} }},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"/json", func(t *testing.T) {
			t.Parallel()
			res := tc.newRes()
			data, err := json.Marshal(res)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			out := tc.newRes()
			// reset to zero by allocating a fresh instance via reflection-free path:
			// simply unmarshal into a freshly constructed value of the same type.
			out = freshOf(res)
			if err := json.Unmarshal(data, out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			got := out.GetMarks()
			if len(got) != len(marks) || got[0] != "critical" || got[1] != "network" {
				t.Errorf("marks round-trip failed: got %v, want %v", got, marks)
			}
		})
		t.Run(tc.name+"/yaml", func(t *testing.T) {
			t.Parallel()
			res := tc.newRes()
			data, err := yaml.Marshal(res)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			out := freshOf(res)
			if err := yaml.Unmarshal(data, out); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			got := out.GetMarks()
			if len(got) != len(marks) || got[0] != "critical" || got[1] != "network" {
				t.Errorf("marks round-trip failed: got %v, want %v", got, marks)
			}
		})
	}
}

// freshOf returns a freshly-constructed value of the same dynamic type as src
// so we can round-trip into a clean target.
func freshOf(src interface{ GetMarks() []string }) interface{ GetMarks() []string } {
	switch src.(type) {
	case *Addr:
		return &Addr{}
	case *Command:
		return &Command{}
	case *DNS:
		return &DNS{}
	case *File:
		return &File{}
	case *Gossfile:
		return &Gossfile{}
	case *Group:
		return &Group{}
	case *HTTP:
		return &HTTP{}
	case *Interface:
		return &Interface{}
	case *KernelParam:
		return &KernelParam{}
	case *Matching:
		return &Matching{}
	case *Mount:
		return &Mount{}
	case *Package:
		return &Package{}
	case *Port:
		return &Port{}
	case *Process:
		return &Process{}
	case *Service:
		return &Service{}
	case *User:
		return &User{}
	default:
		return nil
	}
}

// TestResourceMarksEmptyOmitted ensures the Marks field is omitted from JSON
// output when empty, preserving backward compatibility for existing gossfiles.
func TestResourceMarksEmptyOmitted(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		res  interface{}
	}{
		{"HTTP", &HTTP{Status: 200}},
		{"Command", &Command{ExitStatus: 0}},
		{"File", &File{Exists: true}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tc.res)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if containsString(string(data), "marks") {
				t.Errorf("expected no 'marks' key in JSON output, got: %s", string(data))
			}
		})
	}
}

func containsString(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
