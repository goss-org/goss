package outputs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"sort"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

const (
	discoveryValidationFailure = 1
	discoveryContractFailure   = 2
)

// Discovery emits validation observations as a document that can be loaded by
// a later goss invocation through --vars.
type Discovery struct{}

type discoveryObservation struct {
	ResourceType     string            `json:"resource-type"`
	ResourceID       string            `json:"resource-id"`
	Outcome          string            `json:"outcome"`
	Successful       bool              `json:"successful"`
	Skipped          bool              `json:"skipped"`
	Values           map[string]any    `json:"values"`
	PropertyOutcomes map[string]string `json:"property-outcomes"`
	Errors           map[string]string `json:"errors"`
	RawValues        map[string]any    `json:"raw-values,omitempty"`

	propertyCount int
	skippedCount  int
}

type discoveryProperty struct {
	outcome  string
	value    any
	hasValue bool
	raw      any
	hasRaw   bool
	err      string
	hasError bool
}

type discoveryBuild struct {
	observations map[string]*discoveryObservation
	properties   map[string]map[string]discoveryProperty
	aliases      map[string]discoveryIdentity
	identities   map[discoveryIdentity]string
	excludeRaw   bool
}

type discoveryIdentity struct {
	resourceType string
	resourceID   string
}

type discoveryContractError struct {
	kind     string
	variable string
	property string
	detail   string
}

func (e *discoveryContractError) Error() string {
	location := ""
	if e.variable != "" {
		location = fmt.Sprintf(" for variable %q", e.variable)
	}
	if e.property != "" {
		location += fmt.Sprintf(" property %q", e.property)
	}
	return fmt.Sprintf("%s%s: %s", e.kind, location, e.detail)
}

func (Discovery) ValidOptions() []*formatOption {
	return []*formatOption{
		{name: foPretty},
		{name: foSort},
		{name: foExcludeRaw},
	}
}

// Output consumes the complete result stream before writing so contract
// failures never leave a partial variables document on stdout.
func (d *Discovery) Output(w io.Writer, groups <-chan []resource.TestResult, outConfig util.OutputConfig) int {
	results := collectDiscoveryResults(groups)
	builder := newDiscoveryBuild(util.IsValueInList(foExcludeRaw, outConfig.FormatOptions))
	for _, result := range results {
		if err := builder.add(result); err != nil {
			fmt.Fprintf(os.Stderr, "discovery format: %v\n", err)
			return discoveryContractFailure
		}
	}

	document, exitCode := builder.finish()

	encoded, err := d.encode(document, util.IsValueInList(foPretty, outConfig.FormatOptions))
	if err != nil {
		fmt.Fprintf(os.Stderr, "discovery format: encoding failure: %v\n", err)
		return discoveryContractFailure
	}

	encoded = append(encoded, '\n')
	if err := writeDiscoveryDocument(w, encoded); err != nil {
		fmt.Fprintf(os.Stderr, "discovery format: output failure: %v\n", err)
		return discoveryContractFailure
	}
	return exitCode
}

func (d *Discovery) encode(document any, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(document, "", "  ")
	}
	return json.Marshal(document)
}

func collectDiscoveryResults(groups <-chan []resource.TestResult) []resource.TestResult {
	var results []resource.TestResult
	for group := range groups {
		results = append(results, group...)
	}

	// Validation workers may deliver groups in any order. Sorting before
	// aggregation also makes which conflict is reported independent of arrival.
	sort.SliceStable(results, func(i, j int) bool {
		return discoveryResultKey(results[i]) < discoveryResultKey(results[j])
	})
	return results
}

func discoveryResultKey(result resource.TestResult) string {
	register := ""
	if value, ok := result.Meta["register"]; ok {
		register = fmt.Sprintf("%T:%v", value, value)
	}
	key, _ := json.Marshal([]string{
		result.ResourceType,
		result.ResourceId,
		register,
		result.Property,
		result.ToOutcome(),
	})
	return string(key)
}

func newDiscoveryBuild(excludeRaw bool) *discoveryBuild {
	return &discoveryBuild{
		observations: map[string]*discoveryObservation{},
		properties:   map[string]map[string]discoveryProperty{},
		aliases:      map[string]discoveryIdentity{},
		identities:   map[discoveryIdentity]string{},
		excludeRaw:   excludeRaw,
	}
}

func (b *discoveryBuild) add(result resource.TestResult) error {
	variable, err := discoveryVariableName(result)
	if err != nil {
		return err
	}
	identity := discoveryIdentity{resourceType: result.ResourceType, resourceID: result.ResourceId}

	if previous, ok := b.identities[identity]; ok && previous != variable {
		return &discoveryContractError{
			kind:     "invalid register metadata",
			variable: variable,
			detail:   fmt.Sprintf("resource identity already registered as %q", previous),
		}
	}
	if previous, ok := b.aliases[variable]; ok && previous != identity {
		return &discoveryContractError{
			kind:     "alias collision",
			variable: variable,
			detail:   "multiple resource identities resolve to the same name",
		}
	}
	b.identities[identity] = variable
	b.aliases[variable] = identity

	observation, ok := b.observations[variable]
	if !ok {
		observation = &discoveryObservation{
			ResourceType:     result.ResourceType,
			ResourceID:       result.ResourceId,
			Values:           map[string]any{},
			PropertyOutcomes: map[string]string{},
			Errors:           map[string]string{},
			RawValues:        map[string]any{},
		}
		b.observations[variable] = observation
		b.properties[variable] = map[string]discoveryProperty{}
	}

	property, err := b.property(result, variable)
	if err != nil {
		return err
	}
	if previous, exists := b.properties[variable][result.Property]; exists {
		if reflect.DeepEqual(previous, property) {
			return nil
		}
		return &discoveryContractError{
			kind:     "duplicate-property conflict",
			variable: variable,
			property: result.Property,
			detail:   "repeated results disagree",
		}
	}
	b.properties[variable][result.Property] = property

	observation.propertyCount++
	if property.outcome == resource.OutcomeSkip {
		observation.skippedCount++
	}
	observation.PropertyOutcomes[result.Property] = property.outcome
	if property.hasValue {
		observation.Values[result.Property] = property.value
	}
	if property.hasRaw && !b.excludeRaw {
		observation.RawValues[result.Property] = property.raw
	}
	if property.hasError {
		observation.Errors[result.Property] = property.err
	}
	return nil
}

func discoveryVariableName(result resource.TestResult) (string, error) {
	value, exists := result.Meta["register"]
	if !exists {
		return result.ResourceId, nil
	}
	register, ok := value.(string)
	if !ok || register == "" {
		return "", &discoveryContractError{
			kind:   "invalid register metadata",
			detail: "meta.register must be a non-empty string",
		}
	}
	return register, nil
}

func (b *discoveryBuild) property(result resource.TestResult, variable string) (discoveryProperty, error) {
	property := discoveryProperty{outcome: result.ToOutcome()}
	if result.Err != nil {
		property.err = result.Err.Error()
		property.hasError = true
	}

	if result.Result != resource.SKIP && (result.Err == nil || result.MatcherResult.Actual != nil) {
		value, err := normalizeDiscoveryValue(result.MatcherResult.Actual)
		if err != nil {
			kind := "unsupported value"
			var encodingErr *discoveryEncodingError
			if errors.As(err, &encodingErr) {
				kind = "encoding failure"
			}
			return discoveryProperty{}, &discoveryContractError{
				kind:     kind,
				variable: variable,
				property: result.Property,
				detail:   err.Error(),
			}
		}
		property.value = value
		property.hasValue = true
	}

	if result.Result != resource.SKIP && !b.excludeRaw && len(result.MatcherResult.TransformerChain) > 0 {
		raw, err := normalizeDiscoveryValue(result.MatcherResult.UntransformedValue)
		if err != nil {
			kind := "unsupported value"
			var encodingErr *discoveryEncodingError
			if errors.As(err, &encodingErr) {
				kind = "encoding failure"
			}
			return discoveryProperty{}, &discoveryContractError{
				kind:     kind,
				variable: variable,
				property: result.Property,
				detail:   "raw value: " + err.Error(),
			}
		}
		property.raw = raw
		property.hasRaw = true
	}
	return property, nil
}

func (b *discoveryBuild) finish() (map[string]*discoveryObservation, int) {
	exitCode := 0
	for _, observation := range b.observations {
		observation.Outcome = resource.OutcomePass
		for _, outcome := range observation.PropertyOutcomes {
			if discoveryOutcomeRank(outcome) > discoveryOutcomeRank(observation.Outcome) {
				observation.Outcome = outcome
			}
		}
		observation.Successful = observation.Outcome == resource.OutcomePass
		observation.Skipped = observation.propertyCount > 0 && observation.skippedCount == observation.propertyCount
		if len(observation.RawValues) == 0 || b.excludeRaw {
			observation.RawValues = nil
		}
		if observation.Outcome == resource.OutcomeFail || observation.Outcome == resource.OutcomeUnknown {
			exitCode = discoveryValidationFailure
		}
	}
	return b.observations, exitCode
}

func discoveryOutcomeRank(outcome string) int {
	switch outcome {
	case resource.OutcomeUnknown:
		return 3
	case resource.OutcomeFail:
		return 2
	case resource.OutcomeSkip:
		return 1
	default:
		return 0
	}
}

func normalizeDiscoveryValue(value any) (any, error) {
	return normalizeDiscoveryReflect(reflect.ValueOf(value), map[discoveryVisit]bool{})
}

type discoveryVisit struct {
	typ reflect.Type
	ptr uintptr
}

type discoveryEncodingError struct {
	err error
}

func (e *discoveryEncodingError) Error() string {
	return e.err.Error()
}

func (e *discoveryEncodingError) Unwrap() error {
	return e.err
}

func normalizeDiscoveryReflect(value reflect.Value, seen map[discoveryVisit]bool) (any, error) {
	if !value.IsValid() {
		return nil, nil
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil, nil
		}
		return normalizeDiscoveryReflect(value.Elem(), seen)
	}
	if value.Kind() == reflect.Pointer && value.IsNil() {
		return nil, nil
	}
	if marshaled, handled, err := marshalDiscoveryReflect(value); handled {
		return marshaled, err
	}

	switch value.Kind() {
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.String:
		return value.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint(), nil
	case reflect.Float32, reflect.Float64:
		result := value.Float()
		if math.IsNaN(result) || math.IsInf(result, 0) {
			return nil, fmt.Errorf("%s value must be finite", value.Type())
		}
		return result, nil
	case reflect.Slice:
		if value.IsNil() {
			return nil, nil
		}
		visit := discoveryVisit{typ: value.Type(), ptr: value.Pointer()}
		if seen[visit] {
			return nil, fmt.Errorf("cyclic %s", value.Type())
		}
		seen[visit] = true
		defer delete(seen, visit)
		result := make([]any, value.Len())
		for i := 0; i < value.Len(); i++ {
			item, err := normalizeDiscoveryReflect(value.Index(i), seen)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			result[i] = item
		}
		return result, nil
	case reflect.Array:
		result := make([]any, value.Len())
		for i := 0; i < value.Len(); i++ {
			item, err := normalizeDiscoveryReflect(value.Index(i), seen)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			result[i] = item
		}
		return result, nil
	case reflect.Map:
		if value.IsNil() {
			return nil, nil
		}
		if value.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("map key type %s is not a string", value.Type().Key())
		}
		visit := discoveryVisit{typ: value.Type(), ptr: value.Pointer()}
		if seen[visit] {
			return nil, fmt.Errorf("cyclic %s", value.Type())
		}
		seen[visit] = true
		defer delete(seen, visit)
		result := make(map[string]any, value.Len())
		keys := value.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		for _, key := range keys {
			item, err := normalizeDiscoveryReflect(value.MapIndex(key), seen)
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", key.String(), err)
			}
			result[key.String()] = item
		}
		return result, nil
	default:
		return nil, fmt.Errorf("type %s is not JSON-compatible", value.Type())
	}
}

func marshalDiscoveryReflect(value reflect.Value) (any, bool, error) {
	if value.CanInterface() {
		if marshaler, ok := value.Interface().(json.Marshaler); ok {
			encoded, err := marshalDiscoveryJSON(marshaler, value.Type())
			return encoded, true, err
		}
	}
	// encoding/json also uses pointer-receiver methods on addressable elements,
	// such as T values inside a []T. Map and interface elements are not addressable.
	if value.CanAddr() && value.Addr().CanInterface() {
		if marshaler, ok := value.Addr().Interface().(json.Marshaler); ok {
			encoded, err := marshalDiscoveryJSON(marshaler, value.Addr().Type())
			return encoded, true, err
		}
	}
	return nil, false, nil
}

func marshalDiscoveryJSON(marshaler json.Marshaler, typ reflect.Type) (any, error) {
	encoded, err := marshaler.MarshalJSON()
	if err != nil {
		return nil, &discoveryEncodingError{err: err}
	}
	if !json.Valid(encoded) {
		return nil, &discoveryEncodingError{err: fmt.Errorf("%s returned invalid JSON", typ)}
	}
	return json.RawMessage(append([]byte(nil), encoded...)), nil
}

func writeDiscoveryDocument(w io.Writer, document []byte) error {
	written, err := w.Write(document)
	if err != nil {
		return err
	}
	if written != len(document) {
		return io.ErrShortWrite
	}
	return nil
}

// Ensure the formatter continues to satisfy the output registry contract.
var _ Outputer = (*Discovery)(nil)
