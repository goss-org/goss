package resource

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/goss-org/goss/matchers"
)

const (
	Value = iota
	Values
	Contains
)

const (
	SUCCESS = iota
	FAIL
	SKIP
	UNKNOWN
)

const (
	OutcomePass    = "pass"
	OutcomeFail    = "fail"
	OutcomeSkip    = "skip"
	OutcomeUnknown = "unknown"
)

var humanOutcomes map[int]string = map[int]string{
	UNKNOWN: OutcomeUnknown,
	SUCCESS: OutcomePass,
	FAIL:    OutcomeFail,
	SKIP:    OutcomeSkip,
}

func HumanOutcomes() map[int]string {
	return humanOutcomes
}

type ValidateError string

func (g ValidateError) Error() string { return string(g) }
func toValidateError(err error) *ValidateError {
	if err == nil {
		return nil
	}
	ve := ValidateError(err.Error())
	return &ve
}

type TestResult struct {
	Successful bool `json:"successful" yaml:"successful"`
	Skipped    bool `json:"skipped" yaml:"skipped"`
	// Resource data
	ResourceId   string `json:"resource-id" yaml:"resource-id"`
	ResourceType string `json:"resource-type" yaml:"resource-type"`
	Property     string `json:"property" yaml:"property"`

	// User added info
	Title string `json:"title" yaml:"title"`
	Meta  meta   `json:"meta" yaml:"meta"`

	// Result
	Result        int                    `json:"result" yaml:"result"`
	Err           *ValidateError         `json:"err" yaml:"err"`
	MatcherResult matchers.MatcherResult `json:"matcher-result" yaml:"matcher-result"`
	StartTime     time.Time              `json:"start-time" yaml:"start-time"`
	EndTime       time.Time              `json:"end-time" yaml:"end-time"`
	Duration      time.Duration          `json:"duration" yaml:"duration"`
}

// ToOutcome converts the enum to a human-friendly string.
func (tr TestResult) ToOutcome() string {
	switch tr.Result {
	case SUCCESS:
		return OutcomePass
	case FAIL:
		return OutcomeFail
	case SKIP:
		return OutcomeSkip
	default:
		return OutcomeUnknown
	}
}

func (t TestResult) SortKey() string {
	return fmt.Sprintf("%s:%s", t.ResourceType, t.ResourceId)
}

func skipResult(typeS string, id string, title string, meta meta, property string, startTime time.Time) TestResult {
	endTime := time.Now()
	return TestResult{
		Result:       SKIP,
		Skipped:      true,
		ResourceType: typeS,
		ResourceId:   id,
		Title:        title,
		Meta:         meta,
		Property:     property,
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
	}
}

func ValidateValue(res ResourceRead, property string, expectedValue any, actual any, skip bool) TestResult {
	if f, ok := actual.(func() (io.Reader, error)); ok {
		if _, ok := expectedValue.([]any); !ok {
			actual = func() (string, error) {
				v, err := f()
				if err != nil {
					return "", err
				}
				i, err := matchers.ReaderToString{}.Transform(v)
				if err != nil {
					return "", err
				}
				return i.(string), nil
			}
		}
	}
	return ValidateGomegaValue(res, property, expectedValue, actual, skip)
}

func ValidateGomegaValue(res ResourceRead, property string, expectedValue any, actual any, skip bool) TestResult {
	id := res.ID()
	title := res.GetTitle()
	meta := res.GetMeta()
	typ := reflect.TypeOf(res)
	typeS := strings.Split(typ.String(), ".")[1]
	startTime := time.Now()
	if skip {
		return skipResult(
			typeS,
			id,
			title,
			meta,
			property,
			startTime,
		)
	}

	var foundValue any
	var gomegaMatcher matchers.GossMatcher
	var err error
	switch f := actual.(type) {
	case func() (bool, error):
		foundValue, err = f()
	case func() (string, error):
		foundValue, err = f()
	case func() (int, error):
		foundValue, err = f()
	case func() ([]string, error):
		foundValue, err = f()
	case func() (any, error):
		foundValue, err = f()
	case func() (io.Reader, error):
		foundValue, err = f()
		gomegaMatcher = matchers.HavePatterns(expectedValue)
	default:
		err = fmt.Errorf("Unknown method signature: %t", f)
	}

	var success bool
	if gomegaMatcher == nil && err == nil {
		gomegaMatcher, err = matcherToGomegaMatcher(expectedValue)
	}
	if err != nil {
		endTime := time.Now()
		return TestResult{
			Result:       FAIL,
			ResourceType: typeS,
			ResourceId:   id,
			Title:        title,
			Meta:         meta,
			Property:     property,
			Err:          toValidateError(err),
			StartTime:    startTime,
			EndTime:      endTime,
			Duration:     endTime.Sub(startTime),
		}
	}

	success, err = gomegaMatcher.Match(foundValue)

	var matcherResult matchers.MatcherResult
	result := SUCCESS
	if success {
		matcherResult = matchers.MatcherResult{
			Actual:   foundValue,
			Message:  "matches expectation",
			Expected: expectedValue,
		}
	} else {
		matcherResult = gomegaMatcher.FailureResult(foundValue)
		result = FAIL
	}

	endTime := time.Now()
	return TestResult{
		Result:        result,
		ResourceType:  typeS,
		ResourceId:    id,
		Title:         title,
		Meta:          meta,
		Property:      property,
		MatcherResult: matcherResult,
		Err:           toValidateError(err),
		StartTime:     startTime,
		EndTime:       endTime,
		Duration:      endTime.Sub(startTime),
	}
}
