package matchers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/blang/semver/v4"
)

func TestBeSemverConstraint(t *testing.T) {
	type args struct {
		Constraint any
	}
	tests := []struct {
		name string
		args args
		want GossMatcher
	}{
		{
			name: "sanity",
			args: args{Constraint: "> 1.0.0"},
			want: &BeSemverConstraintMatcher{Constraint: "> 1.0.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BeSemverConstraint(tt.args.Constraint)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBeSemverConstraintMatcher_FailureMessage(t *testing.T) {
	type fields struct {
		Constraint any
	}
	type args struct {
		actual any
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult MatcherResult
	}{
		{
			name:   "string",
			fields: fields{Constraint: "> 1.1.0"},
			args:   args{actual: "1.0.0"},
			wantResult: MatcherResult{
				Actual:   "1.0.0",
				Message:  "to satisfy constraint",
				Expected: "> 1.1.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &BeSemverConstraintMatcher{
				Constraint: tt.fields.Constraint,
			}
			gotResult := matcher.FailureResult(tt.args.actual)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}

func TestBeSemverConstraintMatcher_Match(t *testing.T) {
	type fields struct {
		Constraint any
	}
	type args struct {
		actual any
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
			name:   "pre_release_fail",
			fields: fields{Constraint: ">= 4.0.0"},
			args:   args{actual: []string{"4.0.0-rc1"}},
			want: want{
				success: false,
				err:     false,
			},
		},
		{
			name:   "pre_release_valid",
			fields: fields{Constraint: "< 4.0.0"},
			args:   args{actual: []string{"4.0.0-rc1"}},
			want: want{
				success: true,
				err:     false,
			},
		},
		{
			name:   "invalid_version_starting_with_0",
			fields: fields{Constraint: "> 4.0.0"},
			args:   args{actual: []string{"4.4.019-1"}},
			want: want{
				success:    false,
				err:        true,
				errMessage: "Expected a single or list of semver valid version(s).  Got:\n    <[]string | len:1, cap:1>: [\"4.4.019-1\"]",
			},
		},
		{
			name:   "build_fail",
			fields: fields{Constraint: "> 4.0.0"},
			args:   args{actual: []string{"4.4.019+build+build2"}},
			want: want{
				success:    false,
				err:        true,
				errMessage: "Expected a single or list of semver valid version(s).  Got:\n    <[]string | len:1, cap:1>: [\n        \"4.4.019+build+build2\",\n    ]",
			},
		},
		{
			name:   "build_valid",
			fields: fields{Constraint: "> 4.0.0"},
			args:   args{actual: []string{"4.1.0+build"}},
			want: want{
				success: true,
				err:     false,
			},
		},
		{
			name:   "invalid_actual",
			fields: fields{Constraint: nil},
			args:   args{actual: []string{"4.1.0"}},
			want: want{
				success:    false,
				err:        true,
				errMessage: "Expected a valid semver constraint.  Got:\n    <nil>: nil",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &BeSemverConstraintMatcher{
				Constraint: tt.fields.Constraint,
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

func TestBeSemverConstraintMatcher_NegatedFailureMessage(t *testing.T) {
	type fields struct {
		Constraint any
	}
	type args struct {
		actual any
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult MatcherResult
	}{
		{
			name:   "string",
			fields: fields{Constraint: "> 1.1.0"},
			args:   args{actual: "1.0.0"},
			wantResult: MatcherResult{
				Actual:   "1.0.0",
				Message:  "not to satisfy constraint",
				Expected: "> 1.1.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &BeSemverConstraintMatcher{
				Constraint: tt.fields.Constraint,
			}

			gotResult := matcher.NegatedFailureResult(tt.args.actual)
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}

func Test_toConstraint(t *testing.T) {
	type args struct {
		in any
	}
	type want struct {
		ok bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple",
			args: args{in: "> 1.0.0"},
			want: want{ok: true},
		},
		{
			name: "complex",
			args: args{in: "> 1.0.0 < 2.0.0 || > 4.0.0"},
			want: want{ok: true},
		},
		{
			name: "nil",
			args: args{in: nil},
			want: want{ok: false},
		},
		{
			name: "empty",
			args: args{in: ""},
			want: want{ok: false},
		},
		{
			name: "invalid",
			args: args{in: "invalid"},
			want: want{ok: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConstraint, gotOk := toConstraint(tt.args.in)

			assert.Equal(t, tt.want.ok, gotOk, "success")
			if tt.want.ok {
				assert.NotNil(t, gotConstraint, "constraint shouldn't be nil")
				assert.IsType(t, semver.Range(nil), gotConstraint, "constraint type")
			}
		})
	}
}

func Test_toVersion(t *testing.T) {
	type args struct {
		in any
	}
	type want struct {
		ok bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "simple",
			args: args{in: "1.0.0"},
			want: want{ok: true},
		},
		{
			name: "pre_release",
			args: args{in: "1.2.3-rc1"},
			want: want{ok: true},
		},
		{
			name: "build",
			args: args{in: "1.2.3+build1"},
			want: want{ok: true},
		},
		{
			name: "pre_release_build",
			args: args{in: "1.2.3+build1"},
			want: want{ok: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotOk := toVersion(tt.args.in)

			assert.Equal(t, tt.want.ok, gotOk)
			if tt.want.ok {
				assert.NotNil(t, gotVersion, "version shouldn't be nil")

				if gotVersion != nil {
					assert.Equal(t, fmt.Sprint(tt.args.in), gotVersion.String(), "version")
				}
			}
		})
	}
}

func Test_toVersions(t *testing.T) {
	type args struct {
		in any
	}
	type want struct {
		ok bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "single",
			args: args{in: "1.0.0"},
			want: want{ok: true},
		},
		{
			name: "slice_strings",
			args: args{in: []string{"1.0.0"}},
			want: want{ok: true},
		},
		{
			name: "slice_interfaces",
			args: args{in: []any{"1.0.0"}},
			want: want{ok: true},
		},
		{
			name: "invalid_object",
			args: args{in: want{}},
			want: want{ok: false},
		},
		{
			name: "invalid_object_in_slice",
			args: args{in: []any{want{}}},
			want: want{ok: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersions, gotOk := toVersions(tt.args.in)

			assert.Equal(t, tt.want.ok, gotOk)
			if tt.want.ok {
				assert.NotNil(t, gotVersions, "versions shouldn't be nil")
				assert.NotEmpty(t, gotVersions, "versions shouldn't be empty")

				for i, version := range gotVersions {
					if versions, ok := tt.args.in.([]string); ok {
						assert.Equal(t, fmt.Sprint(versions[i]), version.String())
					} else if versions, ok := tt.args.in.([]any); ok {
						assert.Equal(t, fmt.Sprint(versions[i]), version.String())
					} else {
						assert.Equal(t, fmt.Sprint(tt.args.in), version.String())
					}
				}
			}
		})
	}
}
