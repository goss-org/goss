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

type TestResult struct {
	Successful   bool          `json:"successful" yaml:"successful"`
	ResourceId   string        `json:"resource-id" yaml:"resource-id"`
	ResourceType string        `json:"resource-type" yaml:"resource-type"`
	Title        string        `json:"title" yaml:"title"`
	Meta         meta          `json:"meta" yaml:"meta"`
	TestType     int           `json:"test-type" yaml:"test-type"`
	Property     string        `json:"property" yaml:"property"`
	Err          error         `json:"err" yaml:"err"`
	Expected     []string      `json:"expected" yaml:"expected"`
	Found        []string      `json:"found" yaml:"found"`
	Human        string        `json:"human" yaml:"human"`
	Duration     time.Duration `json:"duration" yaml:"duration"`
}

func ValidateValue(res ResourceRead, property string, expectedValue interface{}, actual interface{}) TestResult {
	id := res.ID()
	title := res.GetTitle()
	meta := res.GetMeta()
	typ := reflect.TypeOf(res)
	typs := strings.Split(typ.String(), ".")[1]
	startTime := time.Now()

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
			ResourceType: typs,
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
	if !success {
		failMessage = gomegaMatcher.FailureMessage(foundValue)
	}

	expected, _ := json.Marshal(expectedValue)
	found, _ := json.Marshal(foundValue)

	return TestResult{
		Successful:   success,
		ResourceType: typs,
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

func newRegexPattern(str string) *regexPattern {
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
	// fixme, don't use MustCompile
	re := regexp.MustCompile(cleanStr)

	return &regexPattern{
		pattern: str,
		re:      re,
		inverse: inverse,
	}

}

func (re *regexPattern) Match(str string) bool {
	return re.re.MatchString(str)
}

func (re *regexPattern) Pattern() string { return re.pattern }
func (re *regexPattern) Inverse() bool   { return re.inverse }

func sliceToPatterns(slice []string) []patternMatcher {
	var patterns []patternMatcher
	for _, s := range slice {
		if (strings.HasPrefix(s, "/") || strings.HasPrefix(s, "!/")) && strings.HasSuffix(s, "/") {
			patterns = append(patterns, newRegexPattern(s))
		} else {
			patterns = append(patterns, newStringPattern(s))
		}
	}
	return patterns
}

func patternsToSlice(patterns []patternMatcher) []string {
	var slice []string
	for _, p := range patterns {
		slice = append(slice, p.Pattern())
	}
	return slice
}

func ValidateContains(res ResourceRead, property string, expectedValues []string, method func() (io.Reader, error)) TestResult {
	id := res.ID()
	title := res.GetTitle()
	meta := res.GetMeta()
	typ := reflect.TypeOf(res)
	typs := strings.Split(typ.String(), ".")[1]
	startTime := time.Now()
	fh, err := method()
	if err != nil {
		return TestResult{
			Successful:   false,
			ResourceType: typs,
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
	notfound := sliceToPatterns(expectedValues)
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
			ResourceType: typs,
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
			ResourceType: typs,
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
		ResourceType: typs,
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
