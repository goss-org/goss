package outputs

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

type formatOption struct {
	name string
}

type Outputer interface {
	Output(io.Writer, <-chan []resource.TestResult, time.Time, util.OutputConfig) int
	ValidOptions() []*formatOption
}

var (
	outputersMu sync.Mutex
	outputers   = map[string]Outputer{
		"documentation": &Documentation{},
		"json_oneline":  &JsonOneline{},
		"json":          &Json{},
		"junit":         &JUnit{},
		"nagios":        &Nagios{},
		"prometheus":    &Prometheus{},
		"rspecish":      &Rspecish{},
		"structured":    &Structured{},
		"tap":           &Tap{},
		"silent":        &Silent{},
	}
	foPerfData = "perfdata"
	foVerbose  = "verbose"
	foPretty   = "pretty"
)

var green = color.New(color.FgGreen).SprintfFunc()
var red = color.New(color.FgRed).SprintfFunc()
var yellow = color.New(color.FgYellow).SprintfFunc()

func humanizeResult(r resource.TestResult) string {
	if r.Err != nil {
		return red("%s: %s: Error: %s", r.ResourceId, r.Property, r.Err)
	}

	switch r.Result {
	case resource.SUCCESS:
		return green("%s: %s: %s: matches expectation: %s", r.ResourceType, r.ResourceId, r.Property, r.Expected)
	case resource.SKIP:
		return yellow("%s: %s: %s: skipped", r.ResourceType, r.ResourceId, r.Property)
	case resource.FAIL:
		if r.Human != "" {
			return red("%s: %s: %s:\n%s", r.ResourceType, r.ResourceId, r.Property, r.Human)
		}
		return humanizeResult2(r)
	default:
		panic(fmt.Sprintf("Unexpected Result Code: %v\n", r.Result))
	}
}

func humanizeResult2(r resource.TestResult) string {
	if r.Err != nil {
		return red("%s: %s: Error: %s", r.ResourceId, r.Property, r.Err)
	}

	switch r.Result {
	case resource.SUCCESS:
		switch r.TestType {
		case resource.Value:
			return green("%s: %s: %s: matches expectation: %s", r.ResourceType, r.ResourceId, r.Property, r.Expected)
		case resource.Values:
			return green("%s: %s: %s: all expectations found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(r.Expected, ", "))
		case resource.Contains:
			return green("%s: %s: %s: all expectations found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(r.Expected, ", "))
		default:
			return red("Unexpected type %d", r.TestType)
		}
	case resource.FAIL:
		switch r.TestType {
		case resource.Value:
			return red("%s: %s: %s: doesn't match, expect: %s found: %s", r.ResourceType, r.ResourceId, r.Property, r.Expected, r.Found)
		case resource.Values:
			return red("%s: %s: %s: expectations not found [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(subtractSlice(r.Expected, r.Found), ", "))
		case resource.Contains:
			return red("%s: %s: %s: patterns not found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(subtractSlice(r.Expected, r.Found), ", "))
		default:
			return red("Unexpected type %d", r.TestType)
		}
	case resource.SKIP:
		return yellow("%s: %s: %s: skipped", r.ResourceType, r.ResourceId, r.Property)
	default:
		panic(fmt.Sprintf("Unexpected Result Code: %v\n", r.Result))
	}
}

func RegisterOutputer(name string, outputer Outputer) {
	outputersMu.Lock()
	defer outputersMu.Unlock()

	if outputer == nil {
		panic("goss: Register outputer is nil")
	}
	if _, dup := outputers[name]; dup {
		panic("goss: Register called twice for ouputer " + name)
	}
	outputers[name] = outputer
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

// FormatOptions returns a sorted list of all the valid options that outputers accept
func FormatOptions() []string {
	outputersMu.Lock()
	defer outputersMu.Unlock()
	found := map[string]*formatOption{}
	for _, o := range outputers {
		for _, opt := range o.ValidOptions() {
			found[opt.name] = opt
		}
	}
	var list []string
	for name := range found {
		list = append(list, name)
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

func GetOutputer(name string) (Outputer, error) {
	if _, ok := outputers[name]; !ok {
		return nil, fmt.Errorf("bad output format: " + name)
	}
	return outputers[name], nil
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

func failedOrSkippedSummary(failedOrSkipped [][]resource.TestResult) string {
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
				s += fmt.Sprintln(humanizeResult(testResult))
			}
			s += fmt.Sprint("\n")
		}
	}
	return s
}
