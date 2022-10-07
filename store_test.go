package goss

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_varsFromString(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "empty_string",
			arg:     ``,
			want:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "empty_JSON",
			arg:     `{}`,
			want:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "JSON_simple",
			arg:  `{"a": "a", "b": 1}`,
			want: map[string]interface{}{
				"a": "a",
				"b": float64(1),
			},
			wantErr: false,
		},
		{
			name: "YAML_simple",
			arg:  `{a: a, b: 1}`,
			want: map[string]interface{}{
				"a": "a",
				"b": 1,
			},
			wantErr: false,
		},
		{
			name: "JSON_float",
			arg:  `{"f": 1.23}`,
			want: map[string]interface{}{
				"f": 1.23,
			},
			wantErr: false,
		},
		{
			name: "YAML_float",
			arg:  `{f: 1.23}`,
			want: map[string]interface{}{
				"f": 1.23,
			},
			wantErr: false,
		},
		{
			name: "JSON_list",
			arg:  `{"l": ["l1", "l2", 3]}`,
			want: map[string]interface{}{
				"l": []interface{}{
					"l1",
					"l2",
					float64(3),
				},
			},
			wantErr: false,
		},
		{
			name: "YAML_list",
			arg:  `{l: [l1, l2, 3]}`,
			want: map[string]interface{}{
				"l": []interface{}{
					"l1",
					"l2",
					3,
				},
			},
			wantErr: false,
		},
		{
			name: "JSON_object",
			arg:  `{"o": {"oa": "a", "oo": { "oo1": 1 } } }`,
			want: map[string]interface{}{
				"o": map[string]interface{}{
					"oa": "a",
					"oo": map[string]interface{}{
						"oo1": float64(1),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "YAML_object",
			arg:  `{o: {oa: a, oo: { oo1: 1 } } }`,
			want: map[string]interface{}{
				"o": map[interface{}]interface{}{
					"oa": "a",
					"oo": map[interface{}]interface{}{
						"oo1": 1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := varsFromString(tt.arg)

			assert.Equal(t, tt.want, got, "map contents")
			assert.Equal(t, tt.wantErr, err != nil, "has error")
		})
	}
}

func Test_loadVars(t *testing.T) {
	fileEmpty, fileEmptyClose := fileMaker(``)
	defer fileEmptyClose()

	fileNil, fileNilClose := fileMaker(``)
	defer fileNilClose()

	fileSimple, fileSimpleClose := fileMaker(`{a: a}`)
	defer fileSimpleClose()

	type args struct {
		varsFile   string
		varsInline string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "both_empty",
			args: args{
				varsFile:   fileEmpty,
				varsInline: `{}`,
			},
			want:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "both_nil",
			args: args{
				varsFile:   fileNil,
				varsInline: `{}`,
			},
			want:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "file_empty",
			args: args{
				varsFile:   fileEmpty,
				varsInline: `{b: b}`,
			},
			want: map[string]interface{}{
				"b": "b",
			},
			wantErr: false,
		},
		{
			name: "inline_empty",
			args: args{
				varsFile:   fileSimple,
				varsInline: `{}`,
			},
			want: map[string]interface{}{
				"a": "a",
			},
			wantErr: false,
		},
		{
			name: "no_overwrite",
			args: args{
				varsFile:   fileSimple,
				varsInline: `{b: b}`,
			},
			want: map[string]interface{}{
				"a": "a",
				"b": "b",
			},
			wantErr: false,
		},
		{
			name: "overwrite",
			args: args{
				varsFile:   fileSimple,
				varsInline: `{a: c, b: b}`,
			},
			want: map[string]interface{}{
				"a": "c",
				"b": "b",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadVars(tt.args.varsFile, tt.args.varsInline)

			assert.Equal(t, tt.want, got, "map contents")
			assert.Equal(t, tt.wantErr, err != nil, "has error")
		})
	}
}

func fileMaker(content string) (string, func()) {
	bytes := []byte(content)

	f, err := os.CreateTemp("", "*")
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}

	return f.Name(), func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
