package matchers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const xmlDoc = `<?xml version="1.0" encoding="utf-8"?>
<order date="2019-02-01">
    <address xmlns:billing="http://localhost/XML/billing">
        <billing:name>Mary Adams</billing:name>
        <billing:city>Old Town</billing:city>
    </address>
    <items>
        <book isbn="9781408845660">
            <title>Harry Potter</title>
            <quantity>1</quantity>
            <price>25</price>
        </book>
        <book isbn="9780544003415">
            <title>Le seigneur des anneaux</title>
            <quantity>1</quantity>
            <price>18</price>
        </book>
    </items>
</order>`

func TestXML(t *testing.T) {
	tests := []struct {
		name string
		args map[string]string
		want GossMatcher
	}{
		{
			name: "sanity",
			args: map[string]string{"xpath": "count(//book)", "result": "2"},
			want: &XMLMatcher{XPath: "count(//book)", ExpectedResult: "2"},
		},
		{
			name: "missing_keys",
			args: map[string]string{},
			want: &XMLMatcher{XPath: "", ExpectedResult: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := XML(tt.args)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestXMLMatcher_Match(t *testing.T) {
	type fields struct {
		XPath          string
		ExpectedResult string
	}
	type want struct {
		success bool
		err     bool
	}
	tests := []struct {
		name   string
		fields fields
		actual any
		want   want
	}{
		// Node queries
		{
			name:   "node_text",
			fields: fields{XPath: "//book[title='Le seigneur des anneaux']/quantity/text()", ExpectedResult: "1"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "node_serialized_with_self",
			fields: fields{XPath: "//book[@isbn='9780544003415']/title", ExpectedResult: "<title>Le seigneur des anneaux</title>"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "several_nodes_are_concatenated",
			fields: fields{XPath: "//book/title", ExpectedResult: "<title>Harry Potter</title><title>Le seigneur des anneaux</title>"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "attribute",
			fields: fields{XPath: "//book[1]/@isbn", ExpectedResult: "<isbn>9781408845660</isbn>"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "namespaced_node",
			fields: fields{XPath: "//address/*[local-name()='city']/text()", ExpectedResult: "Old Town"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "node_mismatch",
			fields: fields{XPath: "//book[1]/quantity/text()", ExpectedResult: "42"},
			actual: xmlDoc,
			want:   want{success: false, err: true},
		},
		// Empty expected result: only node existence is checked
		{
			name:   "existence_found",
			fields: fields{XPath: "//items/book", ExpectedResult: ""},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "existence_not_found",
			fields: fields{XPath: "//items/dvd", ExpectedResult: ""},
			actual: xmlDoc,
			want:   want{success: false},
		},
		{
			name:   "existence_expected_result_is_whitespace_only",
			fields: fields{XPath: "//items/book", ExpectedResult: "   "},
			actual: xmlDoc,
			want:   want{success: true},
		},
		// XPath functions
		{
			name:   "func_count",
			fields: fields{XPath: "count(//book)", ExpectedResult: "2"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "func_sum_float_formatting",
			fields: fields{XPath: "sum(//price)", ExpectedResult: "43"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "func_boolean_true",
			fields: fields{XPath: "boolean(//book[title='Le seigneur des anneaux'])", ExpectedResult: "true"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "func_boolean_false",
			fields: fields{XPath: "boolean(//book[title='Dune'])", ExpectedResult: "false"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "func_string_length",
			fields: fields{XPath: "string-length(//title)", ExpectedResult: "12"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "func_in_predicate_still_yields_nodes",
			fields: fields{XPath: "//book[contains(title, 'anneaux')]/price", ExpectedResult: "<price>18</price>"},
			actual: xmlDoc,
			want:   want{success: true},
		},
		{
			name:   "func_name_inside_a_string_literal_is_not_a_call",
			fields: fields{XPath: "//book[title='count(x)']", ExpectedResult: ""},
			actual: xmlDoc,
			want:   want{success: false},
		},
		{
			name:   "func_uppercase_is_invalid_xpath",
			fields: fields{XPath: "COUNT(//book)", ExpectedResult: "2"},
			actual: xmlDoc,
			want:   want{success: false, err: true},
		},
		{
			name:   "func_mismatch",
			fields: fields{XPath: "count(//book)", ExpectedResult: "3"},
			actual: xmlDoc,
			want:   want{success: false, err: true},
		},
		{
			name:   "func_empty_expected_result_is_not_an_existence_check",
			fields: fields{XPath: "count(//book)", ExpectedResult: ""},
			actual: xmlDoc,
			want:   want{success: false, err: true},
		},
		// Errors
		{
			name:   "actual_is_not_a_string",
			fields: fields{XPath: "//book", ExpectedResult: "1"},
			actual: 42,
			want:   want{success: false, err: true},
		},
		{
			name:   "actual_is_not_xml",
			fields: fields{XPath: "//book", ExpectedResult: "1"},
			actual: "this is definitely not xml <<<",
			want:   want{success: false, err: true},
		},
		{
			name:   "uncompilable_function_expression",
			fields: fields{XPath: "count(//book", ExpectedResult: "2"},
			actual: xmlDoc,
			want:   want{success: false, err: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &XMLMatcher{
				XPath:          tt.fields.XPath,
				ExpectedResult: tt.fields.ExpectedResult,
			}

			gotSuccess, err := matcher.Match(tt.actual)

			assert.Equal(t, tt.want.success, gotSuccess, "has success")
			assert.Equal(t, tt.want.err, err != nil, "has error: %v", err)
		})
	}
}

func TestXMLMatcher_FailureResult(t *testing.T) {
	matcher := &XMLMatcher{XPath: "count(//book)", ExpectedResult: "2"}

	gotResult := matcher.FailureResult("<order/>")

	assert.Equal(t, MatcherResult{
		Actual:   "<order/>",
		Message:  "satisfied XPath match with intended result",
		Expected: "2",
	}, gotResult)
}

func TestXMLMatcher_NegatedFailureResult(t *testing.T) {
	matcher := &XMLMatcher{XPath: "count(//book)", ExpectedResult: "2"}

	gotResult := matcher.NegatedFailureResult("<order/>")

	assert.Equal(t, MatcherResult{
		Actual:   "<order/>",
		Message:  "cannot match xpath query with intended result",
		Expected: "2",
	}, gotResult)
}

func TestXMLMatcher_MarshalJSON(t *testing.T) {
	matcher := &XMLMatcher{XPath: "count(//book)", ExpectedResult: "2"}

	got, err := json.Marshal(matcher)

	assert.NoError(t, err)
	assert.JSONEq(t, `{"xml": "count(//book)"}`, string(got))
}
