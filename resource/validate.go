package resource

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/gomega/types"
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
)

type TestResult struct {
	Successful   bool          `json:"successful" yaml:"successful"`
	ResourceId   string        `json:"resource-id" yaml:"resource-id"`
	ResourceType string        `json:"resource-type" yaml:"resource-type"`
	Title        string        `json:"title" yaml:"title"`
	Meta         meta          `json:"meta" yaml:"meta"`
	TestType     int           `json:"test-type" yaml:"test-type"`
	Result       int           `json:"result" yaml:"result"`
	Property     string        `json:"property" yaml:"property"`
	Err          error         `json:"err" yaml:"err"`
	Expected     []string      `json:"expected" yaml:"expected"`
	Found        []string      `json:"found" yaml:"found"`
	Human        string        `json:"human" yaml:"human"`
	Duration     time.Duration `json:"duration" yaml:"duration"`
}

func skipResult(typeS string, testType int, id string, title string, meta meta, property string, startTime time.Time) TestResult {
	return TestResult{
		Successful:   true,
		Result:       SKIP,
		ResourceType: typeS,
		TestType:     testType,
		ResourceId:   id,
		Title:        title,
		Meta:         meta,
		Property:     property,
		Duration:     startTime.Sub(startTime),
	}
}

func ValidateValue(res ResourceRead, property string, expectedValue interface{}, actual interface{}, skip bool) TestResult {
	id := res.ID()
	title := res.GetTitle()
	meta := res.GetMeta()
	typ := reflect.TypeOf(res)
	typeS := strings.Split(typ.String(), ".")[1]
	startTime := time.Now()
	if skip {
		return skipResult(
			typeS,
			Values,
			id,
			title,
			meta,
			property,
			startTime,
		)
	}

	var foundValue interface{}
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
	default:
		err = fmt.Errorf("Unknown method signature: %t", f)
	}

	expectedValue = sanitizeExpectedValue(expectedValue)
	var gomegaMatcher types.GomegaMatcher
	var success bool
	if err == nil {
		gomegaMatcher, err = matcherToGomegaMatcher(expectedValue)
	}
	if err == nil {
		success, err = gomegaMatcher.Match(foundValue)
	}
	if err != nil {
		return TestResult{
			Successful:   false,
			Result:       FAIL,
			ResourceType: typeS,
			TestType:     Values,
			ResourceId:   id,
			Title:        title,
			Meta:         meta,
			Property:     property,
			Err:          err,
			Duration:     time.Now().Sub(startTime),
		}
	}

	var failMessage string
	var result int
	if !success {
		failMessage = gomegaMatcher.FailureMessage(foundValue)
		result = FAIL
	}

	expected, _ := json.Marshal(expectedValue)
	found, _ := json.Marshal(foundValue)

	return TestResult{
		Successful:   success,
		Result:       result,
		ResourceType: typeS,
		TestType:     Value,
		ResourceId:   id,
		Title:        title,
		Meta:         meta,
		Property:     property,
		Expected:     []string{string(expected)},
		Found:        []string{string(found)},
		Human:        failMessage,
		Err:          err,
		Duration:     time.Now().Sub(startTime),
	}
}

type patternMatcher interface {
	Match(string) bool
	Pattern() string
	Inverse() bool
}

type stringPattern struct {
	pattern      string
	cleanPattern string
	inverse      bool
}

func newStringPattern(str string) *stringPattern {
	var inverse bool
	if strings.HasPrefix(str, "!") {
		inverse = true
	}
	cleanPattern := strings.TrimLeft(str, "\\/!")
	return &stringPattern{
		pattern:      str,
		cleanPattern: cleanPattern,
		inverse:      inverse,
	}
}

func (s *stringPattern) Match(str string) bool {
	return strings.Contains(str, s.cleanPattern)
}

func (s *stringPattern) Pattern() string { return s.pattern }
func (s *stringPattern) Inverse() bool   { return s.inverse }

type regexPattern struct {
	pattern string
	re      *regexp.Regexp
	inverse bool
}

func newRegexPattern(str string) (*regexPattern, error) {
	var inverse bool
	cleanStr := str
	if strings.HasPrefix(str, "!") {
		inverse = true
		cleanStr = cleanStr[1:]
	}
	trimLeft := []rune{'\\', '/'}
	for _, r := range trimLeft {
		if rune(cleanStr[0]) == r {
			cleanStr = cleanStr[1:]
			break
		}
	}
	trimRight := []rune{'/'}
	for _, r := range trimRight {
		if rune(cleanStr[len(cleanStr)-1]) == r {
			cleanStr = cleanStr[:len(cleanStr)-1]
			break
		}
	}

	re, err := regexp.Compile(cleanStr)

	return &regexPattern{
		pattern: str,
		re:      re,
		inverse: inverse,
	}, err

}

func (re *regexPattern) Match(str string) bool {
	return re.re.MatchString(str)
}

func (re *regexPattern) Pattern() string { return re.pattern }
func (re *regexPattern) Inverse() bool   { return re.inverse }

func sliceToPatterns(slice []string) ([]patternMatcher, error) {
	var patterns []patternMatcher
	for _, s := range slice {
		if (strings.HasPrefix(s, "/") || strings.HasPrefix(s, "!/")) && strings.HasSuffix(s, "/") {
			pat, err := newRegexPattern(s)
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, pat)
		} else {
			patterns = append(patterns, newStringPattern(s))
		}
	}
	return patterns, nil
}

func patternsToSlice(patterns []patternMatcher) []string {
	var slice []string
	for _, p := range patterns {
		slice = append(slice, p.Pattern())
	}
	return slice
}

func ValidateContains(res ResourceRead, property string, expectedValues []string, method func() (io.Reader, error), skip bool) TestResult {
	id := res.ID()
	title := res.GetTitle()
	meta := res.GetMeta()
	typ := reflect.TypeOf(res)
	typeS := strings.Split(typ.String(), ".")[1]
	startTime := time.Now()
	if skip {
		return skipResult(
			typeS,
			Values,
			id,
			title,
			meta,
			property,
			startTime,
		)
	}
	var err error
	var fh io.Reader
	var notfound []patternMatcher
	notfound, err = sliceToPatterns(expectedValues)
	// short circuit
	if len(notfound) == 0 && err == nil {
		return TestResult{
			Successful:   true,
			Result:       SUCCESS,
			ResourceType: typeS,
			TestType:     Contains,
			ResourceId:   id,
			Title:        title,
			Meta:         meta,
			Property:     property,
			Expected:     expectedValues,
			Duration:     time.Now().Sub(startTime),
		}
	}
	if err == nil {
		fh, err = method()
	}
	if err != nil {
		return TestResult{
			Successful:   false,
			Result:       FAIL,
			ResourceType: typeS,
			TestType:     Contains,
			ResourceId:   id,
			Title:        title,
			Meta:         meta,
			Property:     property,
			Err:          err,
			Duration:     time.Now().Sub(startTime),
		}
	}
	scanner := bufio.NewScanner(fh)
	var found []patternMatcher
	for scanner.Scan() {
		line := scanner.Text()

		i := 0
		for _, pat := range notfound {
			if pat.Match(line) {
				// Found it, but wasn't supposed to, don't mark it as found, but remove it from search
				if !pat.Inverse() {
					found = append(found, pat)
				}
				continue
			}
			notfound[i] = pat
			i++
		}
		notfound = notfound[:i]
		if len(notfound) == 0 {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return TestResult{
			Successful:   false,
			Result:       FAIL,
			ResourceType: typeS,
			TestType:     Contains,
			ResourceId:   id,
			Title:        title,
			Meta:         meta,
			Property:     property,
			Err:          err,
			Duration:     time.Now().Sub(startTime),
		}
	}

	for _, pat := range notfound {
		// Didn't find it, but we didn't want to.. so we mark it as found
		// Empty pattern should match even if input to scanner is empty
		if pat.Inverse() || pat.Pattern() == "" {
			found = append(found, pat)
		}
	}

	if len(expectedValues) != len(found) {
		return TestResult{
			Successful:   false,
			Result:       FAIL,
			ResourceType: typeS,
			TestType:     Contains,
			ResourceId:   id,
			Title:        title,
			Meta:         meta,
			Property:     property,
			Expected:     expectedValues,
			Found:        patternsToSlice(found),
			Duration:     time.Now().Sub(startTime),
		}
	}
	return TestResult{
		Successful:   true,
		Result:       SUCCESS,
		ResourceType: typeS,
		TestType:     Contains,
		ResourceId:   id,
		Title:        title,
		Meta:         meta,
		Property:     property,
		Expected:     expectedValues,
		Found:        patternsToSlice(found),
		Duration:     time.Now().Sub(startTime),
	}
}
