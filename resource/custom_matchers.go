package resource

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/onsi/gomega/format"
)

type PatternMatcher struct {
	Patterns []string
	NotFound []string
}

func (m *PatternMatcher) Match(actual interface{}) (success bool, err error) {
	fh, ok := actual.(io.Reader)
	if !ok {
		return false, fmt.Errorf("Pattern matcher expects an io.Reader.  Got:\n%s", format.Object(actual, 1))
	}
	var notfound []patternMatcher
	notfound, err = sliceToPatterns(m.Patterns)
	if len(notfound) == 0 && err == nil {
		return true, nil
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
	m.NotFound = subtractSlice(m.Patterns, foundSlice)

	if len(m.Patterns) != len(found) {
		return false, nil
	}
	return true, nil
}

func (m *PatternMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("patterns not found: [%s]", strings.Join(m.NotFound, ", "))
}

func (m *PatternMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("todo: [%s]", strings.Join(m.NotFound, ", "))
}

// TODO: This is duplicated from outputs, refactor
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
