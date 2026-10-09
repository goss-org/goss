package outputs

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goss-org/goss/matchers"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

const (
	wantDiscoveryValidationFailure = 1
	wantDiscoveryContractFailure   = 2
)

func TestIsValidFormat(t *testing.T) {
	if IsValidFormat("ne") {
		t.Fatal("'ne' should not be a valid output format")
	}

	if !IsValidFormat("json") {
		t.Fatal("'json' should be a valid output format")
	}
}

func TestOutputers(t *testing.T) {
	list := Outputers()
	assert.NotEmpty(t, list)
}

func TestGetOutputer(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := GetOutputer("rspecish")
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
	t.Run("not-valid", func(t *testing.T) {
		got, err := GetOutputer("gibberish")
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestOutputFormatOptions(t *testing.T) {
	list := FormatOptions()
	assert.NotEmpty(t, list)

	assert.Contains(t, list, foPerfData)
	assert.Contains(t, list, foPretty)
	assert.Contains(t, list, foVerbose)
	assert.GreaterOrEqual(t, len(list), 4)
}

func TestOptionsRegistration(t *testing.T) {
	registeredOutputs := Outputers()
	assert.Contains(t, registeredOutputs, "documentation")
	assert.Contains(t, registeredOutputs, "json")
	assert.Contains(t, registeredOutputs, "junit")
	assert.Contains(t, registeredOutputs, "nagios")
	assert.Contains(t, registeredOutputs, "prometheus")
	assert.Contains(t, registeredOutputs, "rspecish")
	assert.Contains(t, registeredOutputs, "silent")
	assert.Contains(t, registeredOutputs, "structured")
	assert.Contains(t, registeredOutputs, "tap")
}

func discoveryResult(resourceType, resourceID, property string, outcome int, actual any) resource.TestResult {
	return resource.TestResult{
		Successful:   outcome == resource.SUCCESS,
		Skipped:      outcome == resource.SKIP,
		ResourceId:   resourceID,
		ResourceType: resourceType,
		Property:     property,
		Result:       outcome,
		MatcherResult: matchers.MatcherResult{
			Actual: actual,
		},
	}
}

func discoveryGroups(groups ...[]resource.TestResult) <-chan []resource.TestResult {
	stream := make(chan []resource.TestResult, len(groups))
	for _, group := range groups {
		stream <- group
	}
	close(stream)
	return stream
}

func runDiscovery(t *testing.T, options []string, groups ...[]resource.TestResult) (map[string]any, string, string, int) {
	t.Helper()
	var stdout bytes.Buffer
	outputer, err := GetOutputer("discovery")
	require.NoError(t, err)

	stderrReader, stderrWriter, err := os.Pipe()
	require.NoError(t, err)
	originalStderr := os.Stderr
	os.Stderr = stderrWriter
	defer func() {
		os.Stderr = originalStderr
	}()
	code := outputer.Output(
		&stdout,
		discoveryGroups(groups...),
		util.OutputConfig{FormatOptions: options},
	)
	require.NoError(t, stderrWriter.Close())
	os.Stderr = originalStderr
	diagnostic, err := io.ReadAll(stderrReader)
	require.NoError(t, err)
	require.NoError(t, stderrReader.Close())

	var document map[string]any
	if stdout.Len() > 0 {
		require.NoError(t, json.Unmarshal(stdout.Bytes(), &document))
	}
	return document, stdout.String(), string(diagnostic), code
}

func discoveryVariable(t *testing.T, document map[string]any, name string) map[string]any {
	t.Helper()
	value, ok := document[name]
	require.True(t, ok, "missing discovery variable %q", name)
	variable, ok := value.(map[string]any)
	require.True(t, ok, "discovery variable %q is not an object", name)
	return variable
}

func TestDiscoveryRegistrationOptionsAndEmptyDocument(t *testing.T) {
	assert.Contains(t, Outputers(), "discovery")
	assert.Contains(t, FormatOptions(), foExcludeRaw)
	assert.Contains(t, FormatOptions(), foSort)
	outputer, err := GetOutputer("discovery")
	require.NoError(t, err)

	options := outputer.ValidOptions()
	optionNames := make([]string, len(options))
	for index, option := range options {
		optionNames[index] = option.name
	}
	assert.Contains(t, optionNames, foPretty)
	assert.Contains(t, optionNames, foSort)
	assert.Contains(t, optionNames, foExcludeRaw)

	document, stdout, stderr, code := runDiscovery(t, nil)
	assert.Equal(t, 0, code)
	assert.Equal(t, map[string]any{}, document)
	assert.Equal(t, "{}\n", stdout)
	assert.Empty(t, stderr)
}

func TestDiscoveryPreservesExactRegisterName(t *testing.T) {
	result := discoveryResult("command", "ignored-id", "stdout.value", resource.SUCCESS, "ok")
	result.Meta = map[string]any{"register": " Host.Facts "}
	document, _, _, code := runDiscovery(t, nil, []resource.TestResult{result})
	assert.Equal(t, 0, code)
	assert.Contains(t, document, " Host.Facts ")
	assert.NotContains(t, document, "Host.Facts")
	variable := discoveryVariable(t, document, " Host.Facts ")
	assert.Equal(t, map[string]any{"stdout.value": "ok"}, variable["values"])

	first := discoveryResult("a\x00b", "c", "stdout", resource.SUCCESS, "first")
	first.Meta = map[string]any{"register": "first"}
	second := discoveryResult("a", "b\x00c", "stdout", resource.SUCCESS, "second")
	second.Meta = map[string]any{"register": "second"}
	document, _, stderr, code := runDiscovery(t, nil, []resource.TestResult{first, second})
	assert.Equal(t, 0, code)
	assert.Empty(t, stderr)
	assert.Contains(t, document, "first")
	assert.Contains(t, document, "second")
}

func TestDiscoveryPreservesValuesAndAggregatesRegisteredResource(t *testing.T) {
	zero := discoveryResult("command", "inventory", "zero", resource.SUCCESS, 0)
	zero.Meta = map[string]any{"register": "host_facts"}
	falseValue := discoveryResult("command", "inventory", "false", resource.SUCCESS, false)
	falseValue.Meta = map[string]any{"register": "host_facts"}
	emptyString := discoveryResult("command", "inventory", "empty-string", resource.SUCCESS, "")
	emptyString.Meta = map[string]any{"register": "host_facts"}
	emptyList := discoveryResult("command", "inventory", "empty-list", resource.SUCCESS, []string{})
	emptyList.Meta = map[string]any{"register": "host_facts"}
	emptyMap := discoveryResult("command", "inventory", "empty-map", resource.SUCCESS, map[string]int{})
	emptyMap.Meta = map[string]any{"register": "host_facts"}
	nullValue := discoveryResult("command", "inventory", "null", resource.SUCCESS, nil)
	nullValue.Meta = map[string]any{"register": "host_facts"}
	nested := discoveryResult("command", "inventory", "nested", resource.SUCCESS, map[string]any{
		"ports": []int{80, 443},
		"name":  "web",
	})
	nested.Meta = map[string]any{"register": "host_facts"}

	results := []resource.TestResult{
		nested, emptyMap, zero, nullValue, emptyString, falseValue, emptyList,
	}
	document, stdout, stderr, code := runDiscovery(t, []string{foPretty}, results)

	assert.Equal(t, 0, code)
	assert.Empty(t, stderr)
	assert.True(t, strings.HasSuffix(stdout, "\n"))
	assert.False(t, strings.HasSuffix(stdout, "\n\n"))
	variable := discoveryVariable(t, document, "host_facts")
	assert.Equal(t, "command", variable["resource-type"])
	assert.Equal(t, "inventory", variable["resource-id"])
	assert.Equal(t, resource.OutcomePass, variable["outcome"])
	assert.Equal(t, true, variable["successful"])
	assert.Equal(t, false, variable["skipped"])
	assert.Equal(t, map[string]any{}, variable["errors"])

	values := variable["values"].(map[string]any)
	assert.Equal(t, float64(0), values["zero"])
	assert.Equal(t, false, values["false"])
	assert.Equal(t, "", values["empty-string"])
	assert.Equal(t, []any{}, values["empty-list"])
	assert.Equal(t, map[string]any{}, values["empty-map"])
	assert.Nil(t, values["null"])
	assert.Equal(t, map[string]any{"name": "web", "ports": []any{float64(80), float64(443)}}, values["nested"])
	assert.NotContains(t, variable, "raw-values")

	outcomes := variable["property-outcomes"].(map[string]any)
	for property := range values {
		assert.Equal(t, resource.OutcomePass, outcomes[property])
	}

	compactDocument, compact, _, compactCode := runDiscovery(t, nil, results)
	assert.Equal(t, 0, compactCode)
	assert.Equal(t, document, compactDocument)
	assert.NotEqual(t, stdout, compact)
	_, sorted, _, sortedCode := runDiscovery(t, []string{foSort}, results)
	assert.Equal(t, 0, sortedCode)
	assert.Equal(t, compact, sorted)
}

func TestDiscoveryFallbackFailureSkipErrorAndOutcomePrecedence(t *testing.T) {
	validationErr := resource.ValidateError("permission denied")
	unavailableErr := resource.ValidateError("not available")
	failed := discoveryResult("file", "/etc/app.conf", "mode", resource.FAIL, "0600")
	failed.Err = &validationErr
	unavailable := discoveryResult("file", "/etc/app.conf", "checksum", resource.FAIL, nil)
	unavailable.Err = &unavailableErr
	skipped := discoveryResult("file", "/etc/app.conf", "owner", resource.SKIP, "must-not-leak")
	unknown := discoveryResult("file", "/etc/app.conf", "content", resource.UNKNOWN, nil)

	document, _, stderr, code := runDiscovery(t, nil, []resource.TestResult{skipped, failed, unknown, unavailable})
	assert.Equal(t, wantDiscoveryValidationFailure, code)
	assert.Empty(t, stderr)
	variable := discoveryVariable(t, document, "/etc/app.conf")
	assert.Equal(t, resource.OutcomeUnknown, variable["outcome"])
	assert.Equal(t, false, variable["successful"])
	assert.Equal(t, false, variable["skipped"])
	assert.Equal(t, map[string]any{"checksum": "not available", "mode": "permission denied"}, variable["errors"])
	assert.Equal(t, map[string]any{"content": nil, "mode": "0600"}, variable["values"])
	assert.Equal(t, map[string]any{
		"checksum": resource.OutcomeFail,
		"content":  resource.OutcomeUnknown,
		"mode":     resource.OutcomeFail,
		"owner":    resource.OutcomeSkip,
	}, variable["property-outcomes"])

	allSkipped := discoveryResult("service", "optional", "enabled", resource.SKIP, true)
	document, _, _, code = runDiscovery(t, nil, []resource.TestResult{allSkipped})
	assert.Equal(t, 0, code)
	variable = discoveryVariable(t, document, "optional")
	assert.Equal(t, resource.OutcomeSkip, variable["outcome"])
	assert.Equal(t, false, variable["successful"])
	assert.Equal(t, true, variable["skipped"])
	assert.Equal(t, map[string]any{}, variable["values"])
}

func TestDiscoveryPreservesEmptyErrorAndDetectsPresenceConflict(t *testing.T) {
	emptyMessage := resource.ValidateError("")
	withError := discoveryResult("command", "empty-error", "stdout", resource.FAIL, "value")
	withError.Err = &emptyMessage

	document, stdout, stderr, code := runDiscovery(t, nil, []resource.TestResult{withError})
	assert.Equal(t, wantDiscoveryValidationFailure, code)
	assert.NotEmpty(t, stdout)
	assert.Empty(t, stderr)
	variable := discoveryVariable(t, document, "empty-error")
	assert.Equal(t, map[string]any{"stdout": ""}, variable["errors"])
	assert.Equal(t, map[string]any{"stdout": "value"}, variable["values"])

	withoutError := withError
	withoutError.Err = nil
	_, stdout, stderr, code = runDiscovery(t, nil, []resource.TestResult{withError, withoutError})
	assert.Equal(t, wantDiscoveryContractFailure, code)
	assert.Empty(t, stdout)
	assert.NotEmpty(t, stderr)
}

func TestDiscoveryOutcomePrecedenceCombinations(t *testing.T) {
	tests := []struct {
		name       string
		outcomes   []int
		want       string
		wantCode   int
		allSkipped bool
	}{
		{name: "pass", outcomes: []int{resource.SUCCESS}, want: resource.OutcomePass},
		{name: "pass and skip", outcomes: []int{resource.SUCCESS, resource.SKIP}, want: resource.OutcomeSkip},
		{name: "skip", outcomes: []int{resource.SKIP}, want: resource.OutcomeSkip, allSkipped: true},
		{name: "fail beats skip", outcomes: []int{resource.SKIP, resource.FAIL}, want: resource.OutcomeFail, wantCode: wantDiscoveryValidationFailure},
		{name: "unknown beats fail", outcomes: []int{resource.FAIL, resource.UNKNOWN}, want: resource.OutcomeUnknown, wantCode: wantDiscoveryValidationFailure},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results := make([]resource.TestResult, len(test.outcomes))
			for index, outcome := range test.outcomes {
				results[index] = discoveryResult("command", "aggregate", "property-"+string(rune('a'+index)), outcome, index)
			}
			document, _, _, code := runDiscovery(t, nil, results)
			variable := discoveryVariable(t, document, "aggregate")
			assert.Equal(t, test.wantCode, code)
			assert.Equal(t, test.want, variable["outcome"])
			assert.Equal(t, test.want == resource.OutcomePass, variable["successful"])
			assert.Equal(t, test.allSkipped, variable["skipped"])
		})
	}
}

func TestDiscoveryRawValuesAndExcludeRaw(t *testing.T) {
	transformed := discoveryResult("command", "uptime", "stdout", resource.SUCCESS, float64(42))
	transformed.MatcherResult.TransformerChain = []matchers.Transformer{matchers.ToNumeric{}}
	transformed.MatcherResult.UntransformedValue = " 42 "

	document, _, _, code := runDiscovery(t, nil, []resource.TestResult{transformed})
	assert.Equal(t, 0, code)
	variable := discoveryVariable(t, document, "uptime")
	assert.Equal(t, map[string]any{"stdout": " 42 "}, variable["raw-values"])
	assert.Equal(t, map[string]any{"stdout": float64(42)}, variable["values"])

	failed := transformed
	failed.ResourceId = "failed-transform"
	failed.Result = resource.FAIL
	failed.Successful = false
	document, _, _, code = runDiscovery(t, nil, []resource.TestResult{failed})
	assert.Equal(t, wantDiscoveryValidationFailure, code)
	variable = discoveryVariable(t, document, "failed-transform")
	assert.Equal(t, map[string]any{"stdout": " 42 "}, variable["raw-values"])
	assert.Equal(t, map[string]any{"stdout": float64(42)}, variable["values"])

	document, _, _, code = runDiscovery(t, []string{foExcludeRaw}, []resource.TestResult{transformed})
	assert.Equal(t, 0, code)
	variable = discoveryVariable(t, document, "uptime")
	assert.NotContains(t, variable, "raw-values")
	assert.Equal(t, map[string]any{"stdout": float64(42)}, variable["values"])

	unsupportedRaw := transformed
	unsupportedRaw.ResourceId = "unsupported-raw"
	unsupportedRaw.MatcherResult.UntransformedValue = make(chan int)
	document, _, stderr, code := runDiscovery(t, []string{foExcludeRaw}, []resource.TestResult{unsupportedRaw})
	assert.Equal(t, 0, code)
	assert.Empty(t, stderr)
	variable = discoveryVariable(t, document, "unsupported-raw")
	assert.NotContains(t, variable, "raw-values")

	skippedRaw := transformed
	skippedRaw.ResourceId = "skipped-raw"
	skippedRaw.Result = resource.SKIP
	skippedRaw.Skipped = true
	document, _, stderr, code = runDiscovery(t, nil, []resource.TestResult{skippedRaw})
	assert.Equal(t, 0, code)
	assert.Empty(t, stderr)
	variable = discoveryVariable(t, document, "skipped-raw")
	assert.NotContains(t, variable, "raw-values")
	assert.Equal(t, map[string]any{}, variable["values"])
}

func TestDiscoveryIsDeterministicAcrossGroupAndResultOrder(t *testing.T) {
	a := discoveryResult("command", "a", "stdout", resource.SUCCESS, "alpha")
	b := discoveryResult("command", "b", "stdout", resource.FAIL, "beta")
	c := discoveryResult("command", "a", "exit-status", resource.SUCCESS, 0)

	_, first, firstErr, firstCode := runDiscovery(t, nil,
		[]resource.TestResult{a, b},
		[]resource.TestResult{c},
	)
	_, second, secondErr, secondCode := runDiscovery(t, nil,
		[]resource.TestResult{c},
		[]resource.TestResult{b, a},
	)
	assert.Equal(t, first, second)
	assert.Equal(t, firstErr, secondErr)
	assert.Equal(t, firstCode, secondCode)
	assert.Equal(t, wantDiscoveryValidationFailure, firstCode)

	duplicate := discoveryResult("command", "a", "stdout", resource.SUCCESS, "alpha")
	_, withDuplicate, duplicateErr, duplicateCode := runDiscovery(t, nil,
		[]resource.TestResult{duplicate, a, c, b},
	)
	assert.Equal(t, first, withDuplicate)
	assert.Equal(t, firstErr, duplicateErr)
	assert.Equal(t, firstCode, duplicateCode)
}

func TestDiscoveryRejectsRegisterAndDuplicateConflictsAtomically(t *testing.T) {
	tests := []struct {
		name    string
		results []resource.TestResult
	}{
		{
			name: "empty register",
			results: func() []resource.TestResult {
				result := discoveryResult("command", "a", "stdout", resource.SUCCESS, "a")
				result.Meta = map[string]any{"register": ""}
				return []resource.TestResult{result}
			}(),
		},
		{
			name: "null register",
			results: func() []resource.TestResult {
				result := discoveryResult("command", "a", "stdout", resource.SUCCESS, "a")
				result.Meta = map[string]any{"register": nil}
				return []resource.TestResult{result}
			}(),
		},
		{
			name: "non-string register",
			results: func() []resource.TestResult {
				result := discoveryResult("command", "a", "stdout", resource.SUCCESS, "a")
				result.Meta = map[string]any{"register": 7}
				return []resource.TestResult{result}
			}(),
		},
		{
			name: "alias collision",
			results: func() []resource.TestResult {
				first := discoveryResult("command", "a", "stdout", resource.SUCCESS, "a")
				first.Meta = map[string]any{"register": "shared"}
				second := discoveryResult("command", "b", "stdout", resource.SUCCESS, "b")
				second.Meta = map[string]any{"register": "shared"}
				return []resource.TestResult{first, second}
			}(),
		},
		{
			name: "resource types collide on fallback",
			results: []resource.TestResult{
				discoveryResult("command", "shared", "stdout", resource.SUCCESS, "a"),
				discoveryResult("file", "shared", "mode", resource.SUCCESS, "0600"),
			},
		},
		{
			name: "identity assigned two aliases",
			results: func() []resource.TestResult {
				first := discoveryResult("command", "a", "stdout", resource.SUCCESS, "a")
				first.Meta = map[string]any{"register": "first"}
				second := discoveryResult("command", "a", "exit-status", resource.SUCCESS, 0)
				second.Meta = map[string]any{"register": "second"}
				return []resource.TestResult{first, second}
			}(),
		},
		{
			name: "conflicting duplicate property",
			results: []resource.TestResult{
				discoveryResult("command", "a", "stdout", resource.SUCCESS, "first"),
				discoveryResult("command", "a", "stdout", resource.SUCCESS, "second"),
			},
		},
		{
			name: "conflicting duplicate outcome",
			results: []resource.TestResult{
				discoveryResult("command", "a", "stdout", resource.SUCCESS, "same"),
				discoveryResult("command", "a", "stdout", resource.FAIL, "same"),
			},
		},
		{
			name: "conflicting duplicate raw value",
			results: func() []resource.TestResult {
				first := discoveryResult("command", "a", "stdout", resource.SUCCESS, "same")
				first.MatcherResult.TransformerChain = []matchers.Transformer{matchers.ToString{}}
				first.MatcherResult.UntransformedValue = "first"
				second := first
				second.MatcherResult.UntransformedValue = "second"
				return []resource.TestResult{first, second}
			}(),
		},
		{
			name: "conflicting duplicate error",
			results: func() []resource.TestResult {
				firstError := resource.ValidateError("first")
				secondError := resource.ValidateError("second")
				first := discoveryResult("command", "a", "stdout", resource.FAIL, nil)
				first.Err = &firstError
				second := first
				second.Err = &secondError
				return []resource.TestResult{first, second}
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, stdout, stderr, code := runDiscovery(t, nil, test.results)
			assert.Equal(t, wantDiscoveryContractFailure, code)
			assert.Empty(t, stdout)
			assert.NotEmpty(t, stderr)
		})
	}
}

func TestDiscoveryRejectsUnsupportedValuesAtomically(t *testing.T) {
	cycle := map[string]any{}
	cycle["self"] = cycle
	sliceCycle := []any{nil}
	sliceCycle[0] = sliceCycle
	values := []struct {
		name  string
		value any
	}{
		{name: "non-string map key", value: map[int]string{1: "one"}},
		{name: "structure", value: struct{ Name string }{Name: "no"}},
		{name: "function", value: func() {}},
		{name: "not a number", value: math.NaN()},
		{name: "positive infinity", value: math.Inf(1)},
		{name: "negative infinity", value: math.Inf(-1)},
		{name: "map cycle", value: cycle},
		{name: "slice cycle", value: sliceCycle},
		{name: "unsupported slice child", value: []any{make(chan int)}},
		{name: "unsupported array child", value: [1]any{make(chan int)}},
		{name: "unsupported map child", value: map[string]any{"bad": make(chan int)}},
	}

	for _, test := range values {
		t.Run(test.name, func(t *testing.T) {
			result := discoveryResult("command", "bad", "stdout", resource.SUCCESS, test.value)
			_, stdout, stderr, code := runDiscovery(t, nil, []resource.TestResult{result})
			assert.Equal(t, wantDiscoveryContractFailure, code)
			assert.Empty(t, stdout)
			assert.NotEmpty(t, stderr)
		})
	}

	raw := discoveryResult("command", "bad-raw", "stdout", resource.SUCCESS, "converted")
	raw.MatcherResult.TransformerChain = []matchers.Transformer{matchers.ToString{}}
	raw.MatcherResult.UntransformedValue = make(chan int)
	_, stdout, stderr, code := runDiscovery(t, nil, []resource.TestResult{raw})
	assert.Equal(t, wantDiscoveryContractFailure, code)
	assert.Empty(t, stdout)
	assert.NotEmpty(t, stderr)
}

func TestDiscoveryReportsStableNestedValuePath(t *testing.T) {
	actual := map[string]any{
		"z-later": make(chan int),
		"a-first": []any{map[string]any{"bad": make(chan int)}},
	}
	result := discoveryResult("command", "nested", "stdout", resource.SUCCESS, actual)
	for attempt := 0; attempt < 20; attempt++ {
		_, stdout, stderr, code := runDiscovery(t, nil, []resource.TestResult{result})
		assert.Equal(t, wantDiscoveryContractFailure, code)
		assert.Empty(t, stdout)
		assert.Contains(t, stderr, "a-first")
		assert.NotContains(t, stderr, "z-later")
	}
}

func TestDiscoveryNormalizesSupportedValueBoundaries(t *testing.T) {
	type namedString string
	type namedUint uint16

	var nilSlice []string
	var nilMap map[string]int
	var nilMarshaler *discoveryPointerString
	values := []struct {
		name  string
		input any
		want  any
	}{
		{name: "named string", input: namedString("value"), want: "value"},
		{name: "custom marshaled string", input: discoveryCustomString("original"), want: "custom representation"},
		{name: "unsigned integer", input: namedUint(12), want: float64(12)},
		{name: "finite float", input: float32(1.25), want: float64(1.25)},
		{name: "nil slice", input: nilSlice, want: nil},
		{name: "nil map", input: nilMap, want: nil},
		{name: "nil custom marshaler", input: nilMarshaler, want: nil},
		{name: "array", input: [2]int{1, 2}, want: []any{float64(1), float64(2)}},
		{name: "nested nil interface", input: map[string]any{"empty": nil}, want: map[string]any{"empty": nil}},
	}
	for _, test := range values {
		t.Run(test.name, func(t *testing.T) {
			result := discoveryResult("command", "boundary", "stdout", resource.SUCCESS, test.input)
			document, _, stderr, code := runDiscovery(t, nil, []resource.TestResult{result})
			assert.Equal(t, 0, code)
			assert.Empty(t, stderr)
			variable := discoveryVariable(t, document, "boundary")
			assert.Equal(t, test.want, variable["values"].(map[string]any)["stdout"])
		})
	}
}

type discoveryEncodingFailure string

func (discoveryEncodingFailure) MarshalJSON() ([]byte, error) {
	return nil, errors.New("synthetic encoder failure")
}

type discoveryCustomString string

func (discoveryCustomString) MarshalJSON() ([]byte, error) {
	return []byte(`"custom representation"`), nil
}

type discoveryPointerString string

func (*discoveryPointerString) MarshalJSON() ([]byte, error) {
	return []byte(`"pointer representation"`), nil
}

type discoveryPointerFailure string

func (*discoveryPointerFailure) MarshalJSON() ([]byte, error) {
	return nil, errors.New("pointer encoder failure")
}

func TestDiscoveryAddressablePointerMarshalers(t *testing.T) {
	success := discoveryResult("command", "pointer-success", "stdout", resource.SUCCESS, []discoveryPointerString{"original"})
	document, stdout, stderr, code := runDiscovery(t, nil, []resource.TestResult{success})
	require.Equal(t, 0, code)
	require.Empty(t, stderr)
	require.NotEmpty(t, stdout)
	variable := discoveryVariable(t, document, "pointer-success")
	assert.Equal(t, []any{"pointer representation"}, variable["values"].(map[string]any)["stdout"])

	nilElement := discoveryResult("command", "pointer-nil", "stdout", resource.SUCCESS, []*discoveryPointerString{nil})
	document, _, stderr, code = runDiscovery(t, nil, []resource.TestResult{nilElement})
	require.Equal(t, 0, code)
	require.Empty(t, stderr)
	variable = discoveryVariable(t, document, "pointer-nil")
	assert.Equal(t, []any{nil}, variable["values"].(map[string]any)["stdout"])

	failure := discoveryResult("command", "pointer-failure", "stdout", resource.SUCCESS, []discoveryPointerFailure{"original"})
	_, stdout, stderr, code = runDiscovery(t, nil, []resource.TestResult{success, failure})
	assert.Equal(t, wantDiscoveryContractFailure, code)
	assert.Empty(t, stdout)
	assert.NotEmpty(t, stderr)
}

type discoveryInvalidJSON string

func (discoveryInvalidJSON) MarshalJSON() ([]byte, error) {
	return []byte(`not-json`), nil
}

func TestDiscoveryEncodingFailureIsDiagnosedWithoutOutput(t *testing.T) {
	tests := []struct {
		name   string
		result resource.TestResult
	}{
		{
			name:   "actual marshaler error",
			result: discoveryResult("command", "bad-actual", "stdout", resource.SUCCESS, discoveryEncodingFailure("value")),
		},
		{
			name:   "actual marshaler returns invalid JSON",
			result: discoveryResult("command", "invalid-json", "stdout", resource.SUCCESS, discoveryInvalidJSON("value")),
		},
		{
			name: "raw marshaler error",
			result: func() resource.TestResult {
				result := discoveryResult("command", "bad-raw", "stdout", resource.SUCCESS, "converted")
				result.MatcherResult.TransformerChain = []matchers.Transformer{matchers.ToString{}}
				result.MatcherResult.UntransformedValue = discoveryEncodingFailure("raw")
				return result
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, stdout, stderr, code := runDiscovery(t, nil, []resource.TestResult{test.result})
			assert.Equal(t, wantDiscoveryContractFailure, code)
			assert.Empty(t, stdout)
			assert.NotEmpty(t, stderr)
		})
	}
}

type discoveryShortWriter struct {
	data bytes.Buffer
}

func (w *discoveryShortWriter) Write(data []byte) (int, error) {
	if len(data) > 3 {
		data = data[:3]
	}
	return w.data.Write(data)
}

type discoveryFailWriter struct {
	err error
}

func (w discoveryFailWriter) Write([]byte) (int, error) {
	return 0, w.err
}

type discoveryZeroWriter struct{}

func (discoveryZeroWriter) Write([]byte) (int, error) {
	return 0, nil
}

func TestDiscoveryHandlesShortWritesAndOutputFailures(t *testing.T) {
	result := discoveryResult("command", "ok", "stdout", resource.SUCCESS, "value")
	short := &discoveryShortWriter{}
	outputer, err := GetOutputer("discovery")
	require.NoError(t, err)
	code := outputer.Output(short, discoveryGroups([]resource.TestResult{result}), util.OutputConfig{})
	assert.Equal(t, wantDiscoveryContractFailure, code)
	assert.NotEmpty(t, short.data.Bytes())

	for name, test := range map[string]struct {
		writer io.Writer
	}{
		"explicit error": {writer: discoveryFailWriter{err: errors.New("disk full")}},
		"zero write":     {writer: discoveryZeroWriter{}},
	} {
		t.Run(name, func(t *testing.T) {
			stderrReader, stderrWriter, pipeErr := os.Pipe()
			require.NoError(t, pipeErr)
			originalStderr := os.Stderr
			os.Stderr = stderrWriter
			code := outputer.Output(test.writer, discoveryGroups([]resource.TestResult{result}), util.OutputConfig{})
			require.NoError(t, stderrWriter.Close())
			os.Stderr = originalStderr
			diagnostic, readErr := io.ReadAll(stderrReader)
			require.NoError(t, readErr)
			require.NoError(t, stderrReader.Close())
			assert.Equal(t, wantDiscoveryContractFailure, code)
			assert.NotEmpty(t, diagnostic)
		})
	}
}

// TestOutputExitCodes pins what every outputer returns, because serve turns a
// non-zero code into a 503 at /healthz. Prometheus was missing this once (#992)
// and structured was missing it too, so all of them are pinned at once rather
// than one per incident.
func TestOutputExitCodes(t *testing.T) {
	now := time.Now()
	pass := resource.TestResult{
		Result: resource.SUCCESS, ResourceType: "File", ResourceId: "/tmp",
		Property: "exists", StartTime: now, EndTime: now,
	}
	// a skip is not a failure, so it must not move the exit code on its own
	skip := resource.TestResult{
		Result: resource.SKIP, ResourceType: "File", ResourceId: "/skip",
		Property: "exists", Skipped: true, StartTime: now, EndTime: now,
	}
	fail := resource.TestResult{
		Result: resource.FAIL, ResourceType: "File", ResourceId: "/nope",
		Property: "exists", StartTime: now, EndTime: now,
	}

	cases := map[string]struct {
		outputer Outputer
		failing  int
	}{
		"discovery":     {&Discovery{}, 1},
		"documentation": {Documentation{}, 1},
		"json":          {Json{}, 1},
		"junit":         {JUnit{}, 1},
		"nagios":        {Nagios{}, 2},
		"prometheus":    {Prometheus{}, 1},
		"rspecish":      {Rspecish{}, 1},
		"silent":        {Silent{}, 1},
		"structured":    {Structured{}, 1},
		"tap":           {Tap{}, 1},
	}

	// a new outputer has to make a deliberate choice here rather than default to 0
	assert.Len(t, cases, len(Outputers()))

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// the prometheus outputer accumulates into package-level counters
			defer resetMetrics()

			assert.Equal(t, 0,
				tc.outputer.Output(io.Discard, makeResults(pass, skip), util.OutputConfig{}),
				"a suite with no failures must exit 0")
			assert.Equal(t, tc.failing,
				tc.outputer.Output(io.Discard, makeResults(pass, skip, fail), util.OutputConfig{}),
				"a suite containing a failure must not exit 0")
		})
	}
}
