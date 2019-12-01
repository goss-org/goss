package matchers

import (
	"testing"

	"github.com/onsi/gomega/types"
	"github.com/stretchr/testify/assert"
)

func TestBeVersion(t *testing.T) {
	type args struct {
		comparator string
		compareTo  interface{}
	}
	tests := []struct {
		name string
		args args
		want types.GomegaMatcher
	}{
		{
			name: "empty",
			args: args{},
			want: &BeVersionMatcher{},
		},
		{
			name: "string",
			args: args{
				comparator: "t1",
				compareTo:  "test",
			},
			want: &BeVersionMatcher{
				Comparator: "t1",
				CompareTo:  "test",
			},
		},
		{
			name: "slice",
			args: args{
				comparator: "t1",
				compareTo:  []string{"test", "test2"},
			},
			want: &BeVersionMatcher{
				Comparator: "t1",
				CompareTo:  []string{"test", "test2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BeVersion(tt.args.comparator, tt.args.compareTo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBeVersionMatcher_FailureMessage(t *testing.T) {
	type fields struct {
		Comparator string
		CompareTo  interface{}
	}
	type args struct {
		actual interface{}
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMessage string
	}{
		{
			name: "simple",
			fields: fields{
				Comparator: "<=",
				CompareTo:  "5.5.5",
			},
			args:        args{actual: "6.0.0"},
			wantMessage: "Expected\n    <string>: 6.0.0\nto be <=\n    <string>: 5.5.5",
		},
		{
			name: "slice_interface",
			fields: fields{
				Comparator: "<=",
				CompareTo:  "5.5.5",
			},
			args:        args{actual: []interface{}{"6.0.0"}},
			wantMessage: "Expected\n    <[]interface {} | len:1, cap:1>: [\"6.0.0\"]\nto be <=\n    <string>: 5.5.5",
		},
		{
			name: "slice_string",
			fields: fields{
				Comparator: "<=",
				CompareTo:  "5.5.5",
			},
			args:        args{actual: []string{"6.0.0"}},
			wantMessage: "Expected\n    <[]string | len:1, cap:1>: [\"6.0.0\"]\nto be <=\n    <string>: 5.5.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &BeVersionMatcher{
				Comparator: tt.fields.Comparator,
				CompareTo:  tt.fields.CompareTo,
			}
			gotMessage := matcher.FailureMessage(tt.args.actual)
			assert.Equal(t, tt.wantMessage, gotMessage)
		})
	}
}

func TestBeVersionMatcher_Match(t *testing.T) {
	type fields struct {
		Comparator string
		CompareTo  interface{}
	}
	type args struct {
		actual interface{}
	}
	type want struct {
		success    bool
		err        bool
		errMessage string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "bad_comparator",
			fields: fields{
				Comparator: "!=",
				CompareTo:  nil,
			},
			args: args{actual: ""},
			want: want{
				success:    false,
				err:        true,
				errMessage: "Unknown comparator: !=",
			},
		},
		{
			name: "bad_compareTo",
			fields: fields{
				Comparator: "==",
				CompareTo:  nil,
			},
			args: args{actual: ""},
			want: want{
				success:    false,
				err:        true,
				errMessage: "Expected a version.  Got:\n    <nil>: nil",
			},
		},
		{
			name: "bad_actual",
			fields: fields{
				Comparator: "==",
				CompareTo:  "2",
			},
			args: args{actual: nil},
			want: want{
				success:    false,
				err:        true,
				errMessage: "Expected a version or a list of versions.  Got:\n    <nil>: nil",
			},
		},
		{
			name: "equal",
			fields: fields{
				Comparator: "==",
				CompareTo:  "2",
			},
			args: args{actual: "2.0.0"},
			want: want{
				success: true,
				err:     false,
			},
		},
		{
			name: "bad_equal",
			fields: fields{
				Comparator: "==",
				CompareTo:  "2.0.0-rc1",
			},
			args: args{actual: "2.0.0"},
			want: want{
				success: false,
				err:     false,
			},
		},
		{
			name: "greater_than",
			fields: fields{
				Comparator: ">",
				CompareTo:  "3.9.99",
			},
			args: args{actual: "4.0.0"},
			want: want{
				success: true,
				err:     false,
			},
		},
		{
			name: "bad_greater_than",
			fields: fields{
				Comparator: ">",
				CompareTo:  "3.10.9",
			},
			args: args{actual: "3.7.9"},
			want: want{
				success: false,
				err:     false,
			},
		},
		{
			name: "greater_than_or_equal",
			fields: fields{
				Comparator: ">=",
				CompareTo:  "3",
			},
			args: args{actual: []interface{}{"4.0.0", "3.0.0"}},
			want: want{
				success: true,
				err:     false,
			},
		},
		{
			name: "bad_greater_than_or_equal",
			fields: fields{
				Comparator: ">=",
				CompareTo:  "3",
			},
			args: args{actual: []interface{}{"2.0.0", "3.0.0"}},
			want: want{
				success: false,
				err:     false,
			},
		},
		{
			name: "less_than",
			fields: fields{
				Comparator: "<",
				CompareTo:  "2",
			},
			args: args{actual: []interface{}{"1.0.0"}},
			want: want{
				success: true,
				err:     false,
			},
		},
		{
			name: "bad_less_than",
			fields: fields{
				Comparator: "<",
				CompareTo:  "2",
			},
			args: args{actual: []interface{}{"2.0.0+ubuntu"}},
			want: want{
				success: false,
				err:     false,
			},
		},
		{
			name: "less_than_or_equal",
			fields: fields{
				Comparator: "<=",
				CompareTo:  "1.0.1",
			},
			args: args{actual: []interface{}{"1.0.0", "1.0.1"}},
			want: want{
				success: true,
				err:     false,
			},
		},
		{
			name: "bad_less_than_or_equal",
			fields: fields{
				Comparator: "<=",
				CompareTo:  "0.0.1",
			},
			args: args{actual: []interface{}{"0.0.1", "1.0.1"}},
			want: want{
				success: false,
				err:     false,
			},
		},
		{
			name: "convert_int",
			fields: fields{
				Comparator: "==",
				CompareTo:  1,
			},
			args: args{actual: "1.0.0"},
			want: want{
				success: true,
				err:     false,
			},
		},
		{
			name: "convert_float",
			fields: fields{
				Comparator: "==",
				CompareTo:  1.2,
			},
			args: args{actual: "1.2.0"},
			want: want{
				success: true,
				err:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &BeVersionMatcher{
				Comparator: tt.fields.Comparator,
				CompareTo:  tt.fields.CompareTo,
			}
			gotSuccess, err := matcher.Match(tt.args.actual)

			assert.Equal(t, tt.want.success, gotSuccess, "has success")
			assert.Equal(t, tt.want.err, err != nil, "has error")
			if tt.want.err {
				assert.EqualError(t, err, tt.want.errMessage, "error message")
			}
		})
	}
}

func TestBeVersionMatcher_NegatedFailureMessage(t *testing.T) {
	type fields struct {
		Comparator string
		CompareTo  interface{}
	}
	type args struct {
		actual interface{}
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMessage string
	}{
		{
			name: "simple",
			fields: fields{
				Comparator: "<=",
				CompareTo:  "5.5.5",
			},
			args:        args{actual: "6.0.0"},
			wantMessage: "Expected\n    <string>: 6.0.0\nnot to be <=\n    <string>: 5.5.5",
		},
		{
			name: "slice_interface",
			fields: fields{
				Comparator: "<=",
				CompareTo:  "5.5.5",
			},
			args:        args{actual: []interface{}{"6.0.0"}},
			wantMessage: "Expected\n    <[]interface {} | len:1, cap:1>: [\"6.0.0\"]\nnot to be <=\n    <string>: 5.5.5",
		},
		{
			name: "slice_string",
			fields: fields{
				Comparator: "<=",
				CompareTo:  "5.5.5",
			},
			args:        args{actual: []string{"6.0.0"}},
			wantMessage: "Expected\n    <[]string | len:1, cap:1>: [\"6.0.0\"]\nnot to be <=\n    <string>: 5.5.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &BeVersionMatcher{
				Comparator: tt.fields.Comparator,
				CompareTo:  tt.fields.CompareTo,
			}
			gotMessage := matcher.NegatedFailureMessage(tt.args.actual)
			assert.Equal(t, tt.wantMessage, gotMessage)
		})
	}
}

func Test_toVersion(t *testing.T) {
	type args struct {
		in interface{}
	}
	type want struct {
		version string
		ok      bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "nil",
			args: args{nil},
			want: want{version: "", ok: false},
		},
		{
			name: "empty",
			args: args{in: ""},
			want: want{version: "", ok: false},
		},
		{
			name: "single",
			args: args{in: "1"},
			want: want{version: "1.0.0", ok: true},
		},
		{
			name: "double",
			args: args{in: "1.2"},
			want: want{version: "1.2.0", ok: true},
		},
		{
			name: "triple",
			args: args{in: "1.2.3"},
			want: want{version: "1.2.3", ok: true},
		},
		{
			name: "quadruple",
			args: args{in: "1.2.3.4"},
			want: want{version: "1.2.3.4", ok: true},
		},
		{
			name: "rc",
			args: args{in: "1.2.3-rc1"},
			want: want{version: "1.2.3-rc1", ok: true},
		},
		{
			name: "int",
			args: args{in: 1},
			want: want{version: "1.0.0", ok: true},
		},
		{
			name: "int",
			args: args{in: 1.2},
			want: want{version: "1.2.0", ok: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := toVersion(tt.args.in)

			assert.Equal(t, tt.want.ok, ok, "success")
			if ok {
				assert.Equal(t, tt.want.version, v.String(), "version")
			}
		})
	}
}

func Test_toVersions(t *testing.T) {
	type args struct {
		in interface{}
	}
	type want struct {
		versions []string
		ok       bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty",
			args: args{in: ""},
			want: want{
				versions: nil,
				ok:       false,
			},
		},
		{
			name: "string",
			args: args{in: "2"},
			want: want{
				versions: []string{"2.0.0"},
				ok:       true,
			},
		},
		{
			name: "slice_string",
			args: args{in: []string{"1.2.3-rc1"}},
			want: want{
				versions: []string{"1.2.3-rc1"},
				ok:       true,
			},
		},
		{
			name: "slice_interface",
			args: args{in: []interface{}{"2", "3.2.1"}},
			want: want{
				versions: []string{"2.0.0", "3.2.1"},
				ok:       true,
			},
		},
		{
			name: "map",
			args: args{in: map[int]int{0: 0}},
			want: want{
				versions: []string{},
				ok:       false,
			},
		},
		{
			name: "slice_of_map",
			args: args{in: []interface{}{map[int]int{0: 0}}},
			want: want{
				versions: []string{},
				ok:       false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := toVersions(tt.args.in)

			assert.Equal(t, tt.want.ok, ok, "success")
			if ok {
				assert.Equal(t, len(tt.want.versions), len(v), "versions length")

				for i, want := range tt.want.versions {
					assert.Equal(t, want, v[i].String(), "version")
				}
			}
		})
	}
}
