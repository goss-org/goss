package goss

import (
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
			want:    nil,
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
