package resource

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/matchers"
	"github.com/icza/dyno"
)

const (
	Value = iota
	Values
	Contains
)

const (
	SUCCESS = "SUCCESS"
	FAIL    = "FAIL"
	SKIP    = "SKIP"
)

const (
	maxScanTokenSize = 1024 * 1024
)

type TestResult struct {
	// Resource data
	ResourceId   string `json:"resource-id" yaml:"resource-id"`
	ResourceType string `json:"resource-type" yaml:"resource-type"`
	Property     string `json:"property" yaml:"property"`
	// User added info
	Title string `json:"title" yaml:"title"`
	Meta  meta   `json:"meta" yaml:"meta"`
	// Result
	Result string `json:"result" yaml:"result"`
	Err    error  `json:"err" yaml:"err"`
	// Matches expectation: ...
	Expected string `json:"expected" yaml:"expected"`
	// Used in skip?.. but why?
	Found string `json:"found" yaml:"found"`
	//Human         string                 `json:"human" yaml:"human"`
	MatcherResult matchers.MatcherResult `json:"matcher-result" yaml:"matcher-result"`
	Duration      time.Duration          `json:"duration" yaml:"duration"`
}

func skipResult(typeS string, id string, title string, meta meta, property string, startTime time.Time) TestResult {
	return TestResult{
		Result:       SKIP,
		ResourceType: typeS,
		ResourceId:   id,
		Title:        title,
		Meta:         meta,
		Property:     property,
		Duration:     startTime.Sub(startTime),
	}
}

func ValidateValue(res ResourceRead, property string, expectedValue interface{}, actual interface{}, skip bool) TestResult {
	if f, ok := actual.(func() (io.Reader, error)); ok {
		if _, ok := expectedValue.([]interface{}); !ok {
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

func ValidateGomegaValue(res ResourceRead, property string, expectedValue interface{}, actual interface{}, skip bool) TestResult {
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

	var foundValue interface{}
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
	case func() (interface{}, error):
		foundValue, err = f()
	case func() (io.Reader, error):
		foundValue, err = f()
		gomegaMatcher = matchers.HavePatterns(expectedValue)
	default:
		err = fmt.Errorf("Unknown method signature: %t", f)
	}

	expectedValue = sanitizeExpectedValue(expectedValue)
	var success bool
	if gomegaMatcher == nil && err == nil {
		gomegaMatcher, err = matcherToGomegaMatcher(expectedValue)
	}
	if err == nil {
		success, err = gomegaMatcher.Match(foundValue)
	}
	if err != nil {
		return TestResult{
			Result:       FAIL,
			ResourceType: typeS,
			ResourceId:   id,
			Title:        title,
			Meta:         meta,
			Property:     property,
			Err:          err,
			Duration:     time.Now().Sub(startTime),
		}
	}

	//json := jsoniter.ConfigCompatibleWithStandardLibrary
	expected, _ := json.Marshal(dyno.ConvertMapI2MapS(expectedValue))
	found, _ := json.Marshal(foundValue)

	var matcherResult matchers.MatcherResult
	result := SUCCESS
	if success {
		matcherResult = matchers.MatcherResult{
			Message:  "matches expectation",
			Expected: expectedValue,
		}
	} else {
		matcherResult = gomegaMatcher.FailureResult(foundValue)
		result = FAIL
	}

	return TestResult{
		Result:        result,
		ResourceType:  typeS,
		ResourceId:    id,
		Title:         title,
		Meta:          meta,
		Property:      property,
		Expected:      string(expected),
		Found:         string(found),
		MatcherResult: matcherResult,
		Err:           err,
		Duration:      time.Now().Sub(startTime),
	}
}
