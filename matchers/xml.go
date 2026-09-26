package matchers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
)

type XMLMatcher struct {
	fakeOmegaMatcher

	XPath          string
	ExpectedResult string
}

func XML(dict map[string]string) GossMatcher {
	return &XMLMatcher{
		XPath:          dict["xpath"],
		ExpectedResult: dict["result"],
	}
}

func (m *XMLMatcher) Match(actual any) (bool, error) {
	xmlStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("XML expect pattern to be a string: \n%s", actual)
	}

	doc, err := xmlquery.Parse(strings.NewReader(xmlStr))
	if err != nil {
		return false, fmt.Errorf("Cannot parse XML string %q: %w", xmlStr, err)
	}

	xp, err := xpath.Compile(m.XPath)
	if err != nil {
		return false, fmt.Errorf("Invalid XPath query %q: %w", m.XPath, err)
	}

	nav := xmlquery.CreateXPathNavigator(doc)

	var strV string
	switch v := xp.Evaluate(nav).(type) {
	case *xpath.NodeIterator:
		nodes := xmlquery.QuerySelectorAll(doc, xp)

		if strings.TrimSpace(m.ExpectedResult) == "" {
			return len(nodes) > 0, nil
		}

		var sb strings.Builder
		for _, n := range nodes {
			sb.WriteString(n.OutputXMLWithOptions(xmlquery.WithEmptyTagSupport(), xmlquery.WithOutputSelf()))
		}
		strV = sb.String()
	case float64:
		strV = strconv.FormatFloat(v, 'G', -1, 64)
	case bool:
		strV = strconv.FormatBool(v)
	case string:
		strV = v
	default:
		return false, fmt.Errorf("Unsupported XPath result type: %T", v)
	}

	if strV == m.ExpectedResult {
		return true, nil
	}
	return false, fmt.Errorf("Cannot match XPath query with attended result; result: %q; expected: %q", strV, m.ExpectedResult)
}

func (m *XMLMatcher) FailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "satisfied XPath match with intended result",
		Expected: m.ExpectedResult,
	}
}

func (m *XMLMatcher) NegatedFailureResult(actual any) MatcherResult {
	return MatcherResult{
		Actual:   actual,
		Message:  "cannot match xpath query with intended result",
		Expected: m.ExpectedResult,
	}
}

func (m *XMLMatcher) MarshalJSON() ([]byte, error) {
	j := make(map[string]any)
	j["xml"] = m.XPath
	return json.Marshal(j)
}
