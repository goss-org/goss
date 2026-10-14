package matchers

import (
	"encoding/json"

	"github.com/onsi/gomega/matchers"
	"github.com/samber/lo"
)

type ConsistOfMatcher struct {
	matchers.ConsistOfMatcher
}

func ConsistOf(elements ...any) GossMatcher {
	return &ConsistOfMatcher{
		matchers.ConsistOfMatcher{
			Elements: elements,
		},
	}
}

func (m *ConsistOfMatcher) FailureResult(actual any) MatcherResult {
	missingElements := getUnexported(m, "missingElements")
	extraElements := getUnexported(m, "extraElements")
	missingEl, ok := missingElements.([]any)
	var foundElements any
	if ok {
		foundElements, _ = lo.Difference(m.Elements, missingEl)
	}
	return MatcherResult{
		Actual:          actual,
		Message:         "to consist of",
		Expected:        m.Elements,
		MissingElements: missingElements,
		ExtraElements:   extraElements,
		FoundElements:   foundElements,
	}
}

func (m *ConsistOfMatcher) NegatedFailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "not to consist of",
		Expected: m.Elements,
	}
}

func (m *ConsistOfMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]any)
	j["consist-of"] = m.Elements
	return json.Marshal(j)
}
