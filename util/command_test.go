package util

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestExecCommandUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ExecCommand
		wantErr bool
	}{
		{
			name:  "string is shell style",
			input: "echo hello world",
			want:  ExecCommand{CmdStr: "echo hello world"},
		},
		{
			name:  "flow sequence is exec style",
			input: "[/bin/echo, hello world]",
			want:  ExecCommand{CmdSlice: []string{"/bin/echo", "hello world"}},
		},
		{
			name:  "block sequence is exec style",
			input: "\n- /bin/echo\n- hello world\n",
			want:  ExecCommand{CmdSlice: []string{"/bin/echo", "hello world"}},
		},
		{
			name:  "bool scalar decodes to its string form (exec: true)",
			input: "true",
			want:  ExecCommand{CmdStr: "true"},
		},
		{
			name:  "null is empty",
			input: "null",
			want:  ExecCommand{},
		},
		{
			name:    "mapping is invalid",
			input:   "{a: b}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ExecCommand
			err := yaml.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestExecCommandUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ExecCommand
		wantErr bool
	}{
		{
			name:  "string is shell style",
			input: `"echo hello world"`,
			want:  ExecCommand{CmdStr: "echo hello world"},
		},
		{
			name:  "array is exec style",
			input: `["/bin/echo", "hello world"]`,
			want:  ExecCommand{CmdSlice: []string{"/bin/echo", "hello world"}},
		},
		{
			name:    "number is invalid",
			input:   `5`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ExecCommand
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestExecCommandMarshalRoundTrip(t *testing.T) {
	shell := ExecCommand{CmdStr: "echo hi"}
	execStyle := ExecCommand{CmdSlice: []string{"/bin/echo", "hi"}}

	for _, want := range []ExecCommand{shell, execStyle} {
		yb, err := yaml.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		var gotY ExecCommand
		if err := yaml.Unmarshal(yb, &gotY); err != nil {
			t.Fatalf("yaml round trip %v: %v", want, err)
		}
		if !reflect.DeepEqual(gotY, want) {
			t.Fatalf("yaml round trip: got %+v, want %+v (marshaled: %q)", gotY, want, yb)
		}

		jb, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		var gotJ ExecCommand
		if err := json.Unmarshal(jb, &gotJ); err != nil {
			t.Fatalf("json round trip %v: %v", want, err)
		}
		if !reflect.DeepEqual(gotJ, want) {
			t.Fatalf("json round trip: got %+v, want %+v (marshaled: %q)", gotJ, want, jb)
		}
	}
}

// holder mirrors how resource.Command embeds ExecCommand with omitempty so we
// can assert an unset command is omitted from serialized output in both formats.
type execHolder struct {
	Exec *ExecCommand `yaml:"exec,omitempty" json:"exec,omitempty"`
	Name string       `yaml:"name,omitempty" json:"name,omitempty"`
}

func TestExecCommandOmittedWhenUnset(t *testing.T) {
	yb, err := yaml.Marshal(execHolder{Name: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if string(yb) != "name: x\n" {
		t.Fatalf("expected exec omitted from yaml, got %q", yb)
	}

	jb, err := json.Marshal(execHolder{Name: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if string(jb) != `{"name":"x"}` {
		t.Fatalf("expected exec omitted from json, got %q", jb)
	}
}
