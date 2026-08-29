package resource

import (
	"reflect"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestCommandGetExec(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  any
	}{
		{
			name:  "no exec falls back to the id",
			input: "foo:\n  exit-status: 0\n",
			want:  "foo",
		},
		{
			name:  "string exec is shell style",
			input: "foo:\n  exit-status: 0\n  exec: echo hi\n",
			want:  "echo hi",
		},
		{
			name:  "list exec is exec style",
			input: "foo:\n  exit-status: 0\n  exec: [/bin/echo, hello world]\n",
			want:  []string{"/bin/echo", "hello world"},
		},
		{
			name:  "bool scalar exec keeps legacy string behavior",
			input: "foo:\n  exit-status: 0\n  exec: true\n",
			want:  "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m CommandMap
			if err := yaml.Unmarshal([]byte(tt.input), &m); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			cmd, ok := m["foo"]
			if !ok {
				t.Fatalf("expected foo resource, got %+v", m)
			}
			got := cmd.GetExec()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GetExec() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCommandInvalidExec(t *testing.T) {
	input := "foo:\n  exit-status: 0\n  exec: {a: b}\n"
	var m CommandMap
	if err := yaml.Unmarshal([]byte(input), &m); err == nil {
		t.Fatalf("expected error for mapping exec, got %+v", m)
	}
}
