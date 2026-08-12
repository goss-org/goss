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

func (m *XMLMatcher) Match(actual any) (success bool, err error) {
	xmlStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("XML expect pattern to be a string: \n%s", actual)
	}

	doc, err := xmlquery.Parse(strings.NewReader(xmlStr))
	if err != nil {
		return false, fmt.Errorf("Cannot parse XML string: \n%s", xmlStr)
	}
	nav := xmlquery.CreateXPathNavigator(doc)

	var strV string
	if hasFunc(m.XPath) {
		xp, err := xpath.Compile(m.XPath)
		if err != nil {
			return false, fmt.Errorf("Cannot compile XPath query: \n%q", m.XPath)
		}
		val := xp.Evaluate(nav)

		switch v := val.(type) {
		case float64:
			strV = strconv.FormatFloat(v, 'G', -1, 64)
		case bool:
			strV = strconv.FormatBool(v)
		case string:
			strV = v
		default:
			return false, fmt.Errorf("unsupported function result type: %T", v)
		}
	} else {
		nodes := xmlquery.Find(doc, m.XPath)

		if strings.TrimSpace(m.ExpectedResult) == "" {
			return len(nodes) > 0, nil
		}

		var sb strings.Builder
		for _, n := range nodes {
			sb.WriteString(n.OutputXMLWithOptions(xmlquery.WithEmptyTagSupport(), xmlquery.WithOutputSelf()))
		}
		strV = sb.String()
	}

	if strV == m.ExpectedResult {
		return true, nil
	}
	return false, fmt.Errorf("Cannot match XPath query with attended result; result: %q; expected: %q", strV, m.ExpectedResult)
}

func hasFunc(expr string) bool {
	// Functions supported by XPath
	// Ref: https://github.com/antchfx/xpath?tab=readme-ov-file#supported-features
	known := []string{
		"boolean", "ceiling", "choose", "concat", "contains", "count", "current", "document", "element-available", "false",
		"floor", "format-number", "function-available", "generate-id", "id", "key", "lang", "last", "local-name", "name", "namespace-uri",
		"normalize-space", "not", "number", "position", "round", "starts-with", "string", "string-length", "substring", "substring-after",
		"substring-before", "sum", "system-property", "translate", "true", "unparsed-entity-url",
	}
	lc := strings.ToLower(expr)
	for _, fn := range known {
		if strings.Contains(lc, fn+"(") {
			return true
		}
	}
	return false
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
