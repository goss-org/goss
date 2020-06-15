package outputs

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aelsabbahy/goss/matchers"
	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
	"github.com/fatih/color"
)

type Outputer interface {
	Output(io.Writer, <-chan []resource.TestResult, time.Time, util.OutputConfig) int
}

var green = color.New(color.FgGreen).SprintfFunc()
var red = color.New(color.FgRed).SprintfFunc()
var yellow = color.New(color.FgYellow).SprintfFunc()

func humanizeResult(r resource.TestResult, compact bool, includeRaw bool) string {
	sep := "\n"
	if compact {
		sep = " "
	}
	if r.Err != nil {
		return red("%s: %s: Error: %s", r.ResourceId, r.Property, r.Err)
	}

	switch r.Result {
	case resource.SUCCESS:
		return green("%s: %s: %s: %s: %s", r.ResourceType, r.ResourceId, r.Property, r.MatcherResult.Message, prettyPrint(r.MatcherResult.Expected, false))
	case resource.FAIL:
		matcherResult := prettyPrintMatcherResult(r.MatcherResult, compact, includeRaw)
		return red("%s: %s: %s:%s%s", r.ResourceType, r.ResourceId, r.Property, sep, matcherResult)
	case resource.SKIP:
		return yellow("%s: %s: %s: skipped", r.ResourceType, r.ResourceId, r.Property)
	default:
		panic(fmt.Sprintf("Unexpected Result Code: %v\n", r.Result))
	}
}

func prettyPrintMatcherResult(m matchers.MatcherResult, compact bool, includeRaw bool) string {
	sep := "\n"
	if compact {
		sep = " "
	}
	var ss []string
	//var s string
	ss = append(ss, "Expected")
	ss = append(ss, prettyPrint(m.Actual, !compact))
	ss = append(ss, m.Message)
	ss = append(ss, prettyPrint(m.Expected, !compact))

	if m.MissingElements != nil {
		ss = append(ss, "the missing elements were")
		ss = append(ss, prettyPrint(m.MissingElements, !compact))
	}
	if m.ExtraElements != nil {
		ss = append(ss, "the extra elements were")
		ss = append(ss, prettyPrint(m.MissingElements, !compact))
	}
	if len(m.TransformerChain) != 0 {
		ss = append(ss, "the transform chain was")
		ss = append(ss, prettyPrint(m.TransformerChain, !compact))
		if includeRaw {
			ss = append(ss, "the raw value was")
			ss = append(ss, prettyPrint(m.UntransformedValue, !compact))
		}
	}
	return strings.Join(ss, sep)
	//return s
}

func prettyPrint(i interface{}, indent bool) string {
	// fixme: error handling
	b, err := json.Marshal(i)
	if err != nil {
		b = []byte(fmt.Sprint(err))
	}
	if indent {
		return indentLines(string(b))
	} else {
		return string(b)
	}
}

// indents a block of text with an indent string
func indentLines(text string) string {
	indent := "    "
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result[:len(result)-1]
}

//func humanizeResult2(r resource.TestResult) string {
//	if r.Err != nil {
//		return red("%s: %s: Error: %s", r.ResourceId, r.Property, r.Err)
//	}
//
//	switch r.Result {
//	case resource.SUCCESS:
//		switch r.TestType {
//		case resource.Value:
//			return green("%s: %s: %s: matches expectation: %s", r.ResourceType, r.ResourceId, r.Property, r.Expected)
//		case resource.Values:
//			return green("%s: %s: %s: all expectations found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(r.Expected, ", "))
//		case resource.Contains:
//			return green("%s: %s: %s: all expectations found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(r.Expected, ", "))
//		default:
//			return red("Unexpected type %d", r.TestType)
//		}
//	case resource.FAIL:
//		switch r.TestType {
//		case resource.Value:
//			return red("%s: %s: %s: doesn't match, expect: %s found: %s", r.ResourceType, r.ResourceId, r.Property, r.Expected, r.Found)
//		case resource.Values:
//			return red("%s: %s: %s: expectations not found [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(subtractSlice(r.Expected, r.Found), ", "))
//		case resource.Contains:
//			return red("%s: %s: %s: patterns not found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(subtractSlice(r.Expected, r.Found), ", "))
//		default:
//			return red("Unexpected type %d", r.TestType)
//		}
//	case resource.SKIP:
//		return yellow("%s: %s: %s: skipped", r.ResourceType, r.ResourceId, r.Property)
//	default:
//		panic(fmt.Sprintf("Unexpected Result Code: %v\n", r.Result))
//	}
//}
//
// Copied from database/sql
var (
	outputersMu           sync.Mutex
	outputers             = make(map[string]Outputer)
	outputerFormatOptions = make(map[string][]string)
)

func RegisterOutputer(name string, outputer Outputer, formatOptions []string) {
	outputersMu.Lock()
	defer outputersMu.Unlock()

	if outputer == nil {
		panic("goss: Register outputer is nil")
	}
	if _, dup := outputers[name]; dup {
		panic("goss: Register called twice for ouputer " + name)
	}
	outputers[name] = outputer
	outputerFormatOptions[name] = formatOptions
}

// Outputers returns a sorted list of the names of the registered outputers.
func Outputers() []string {
	outputersMu.Lock()
	defer outputersMu.Unlock()
	var list []string
	for name := range outputers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// FormatOptions are all the valid options formatters accept
func FormatOptions() []string {
	outputersMu.Lock()
	defer outputersMu.Unlock()
	var list []string
	for _, formatOptions := range outputerFormatOptions {
		for _, opt := range formatOptions {
			if !(util.IsValueInList(opt, list)) {
				list = append(list, opt)
			}
		}
	}
	sort.Strings(list)
	return list
}

// IsValidFormat determines if f is a valid format name based on Outputers()
func IsValidFormat(f string) bool {
	for _, o := range Outputers() {
		if o == f {
			return true
		}
	}

	return false
}

// IsValidFormatOption determines if o is a valid format option based on FormatOptions()
func IsValidFormatOption(o string) bool {
	for _, p := range FormatOptions() {
		if p == o {
			return true
		}
	}

	return false
}

func GetOutputer(name string) (Outputer, error) {
	if _, ok := outputers[name]; !ok {
		return nil, fmt.Errorf("bad output format: " + name)
	}
	return outputers[name], nil
}

//func subtractSlice(x, y []string) []string {
//	m := make(map[string]bool)
//
//	for _, y := range y {
//		m[y] = true
//	}
//
//	var ret []string
//	for _, x := range x {
//		if m[x] {
//			continue
//		}
//		ret = append(ret, x)
//	}
//
//	return ret
//}
//
func header(t resource.TestResult) string {
	var out string
	if t.Title != "" {
		out += fmt.Sprintf("Title: %s\n", t.Title)
	}
	if t.Meta != nil {
		var keys []string
		for k := range t.Meta {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		out += "Meta:\n"
		for _, k := range keys {
			out += fmt.Sprintf("    %v: %v\n", k, t.Meta[k])
		}
	}
	return out
}

func summary(startTime time.Time, count, failed, skipped int) string {
	var s string
	s += fmt.Sprintf("Total Duration: %.3fs\n", time.Since(startTime).Seconds())
	f := green
	if failed > 0 {
		f = red
	}
	s += f("Count: %d, Failed: %d, Skipped: %d\n", count, failed, skipped)
	return s
}

func failedOrSkippedSummary(failedOrSkipped [][]resource.TestResult, includeRaw bool) string {
	var s string
	if len(failedOrSkipped) > 0 {
		s += fmt.Sprint("Failures/Skipped:\n\n")
		for _, failedGroup := range failedOrSkipped {
			first := failedGroup[0]
			header := header(first)
			if header != "" {
				s += fmt.Sprint(header)
			}
			for _, testResult := range failedGroup {
				s += fmt.Sprintln(humanizeResult(testResult, false, includeRaw))
			}
			s += fmt.Sprint("\n")
		}
	}
	return s
}
