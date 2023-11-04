package matchers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/onsi/gomega/format"
)

const (
	maxScanTokenSize = 1024 * 1024
)

type HavePatternsMatcher struct {
	fakeOmegaMatcher

	Elements        interface{}
	missingElements []string
	foundElements   []string
}

func HavePatterns(elements interface{}) GossMatcher {
	return &HavePatternsMatcher{
		Elements: elements,
	}
}

func (m *HavePatternsMatcher) Match(actual interface{}) (success bool, err error) {
	t, ok := m.Elements.([]interface{})
	if !ok {
		return false, fmt.Errorf("HavePatterns matcher expects an array of matchers.  Got:\n%s", format.Object(m.Elements, 1))
	}
	elements := make([]string, len(t))
	for i, v := range t {
		switch v := v.(type) {
		case string:
			elements[i] = v
		default:
			return false, fmt.Errorf("HavePatterns matcher expects patterns to be a string. got: \n%s", format.Object(v, 1))
		}
	}
	notfound, err := sliceToPatterns(elements)
	if err != nil {
		return false, err
	}
	// short circuit
	if len(notfound) == 0 {
		return true, nil
	}
	var fh io.Reader
	switch av := actual.(type) {
	case io.Reader:
		fh = av
	case string:
		fh = strings.NewReader(av)
	case []string:
		fh = strings.NewReader(strings.Join(av, "\n"))
	default:
		err = fmt.Errorf("Incorrect type %T", actual)

	}
	if err != nil {
		return false, err
	}

	defer func() {
		if rc, ok := fh.(io.ReadCloser); ok {
			rc.Close()
		}
	}()

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(nil, maxScanTokenSize)
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
		return false, err
	}

	for _, pat := range notfound {
		// Didn't find it, but we didn't want to.. so we mark it as found
		// Empty pattern should match even if input to scanner is empty
		if pat.Inverse() || pat.Pattern() == "" {
			found = append(found, pat)
		}
	}

	foundSlice := patternsToSlice(found)
	m.foundElements = foundSlice
	if len(elements) != len(found) {
		m.missingElements = subtractSlice(elements, foundSlice)
		return false, nil
	}
	return true, nil
}

func (m *HavePatternsMatcher) FailureResult(actual interface{}) MatcherResult {
	var a interface{}
	switch actual.(type) {
	case string, []string:
		a = actual
	default:
		a = fmt.Sprintf("object: %T", actual)
	}
	return MatcherResult{
		Actual:          a,
		Message:         "to have patterns",
		Expected:        m.Elements,
		MissingElements: m.missingElements,
		FoundElements:   m.foundElements,
	}
}

func (m *HavePatternsMatcher) NegatedFailureResult(actual interface{}) MatcherResult {
	a, ok := actual.(string)
	if !ok {
		a = fmt.Sprintf("object: %T", actual)
	}
	return MatcherResult{
		Actual:   a,
		Message:  "not to have patterns",
		Expected: m.Elements,
	}
}

func appendMissingStrings(message string, missingElements []string) string {
	if len(missingElements) == 0 {
		return message
	}
	return fmt.Sprintf("%s\nthe missing elements were\n%s", message,
		format.Object(missingElements, 1))
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
func subtractSlice(x, y []string) []string {
	m := make(map[string]bool)

	for _, y := range y {
		m[y] = true
	}

	var ret []string
	for _, x := range x {
		if m[x] {
			continue
		}
		ret = append(ret, x)
	}

	return ret
}

func (matcher *HavePatternsMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["have-patterns"] = matcher.Elements
	return json.Marshal(j)
}
