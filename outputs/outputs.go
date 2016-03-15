package outputs

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/fatih/color"
)

type Outputer interface {
	Output(<-chan []resource.TestResult, time.Time) int
}

var green = color.New(color.FgGreen).SprintfFunc()
var red = color.New(color.FgRed).SprintfFunc()

func humanizeResult(r resource.TestResult) string {
	if r.Err != nil {
		return red("%s: %s: Error: %s", r.ResourceId, r.Property, r.Err)
	}

	if r.Successful {
		return green("%s: %s: %s: matches expectation: %s", r.ResourceType, r.ResourceId, r.Property, r.Expected)
	} else {
		if r.Human != "" {
			return red("%s: %s: %s:\n%s", r.ResourceType, r.ResourceId, r.Property, r.Human)
		}
		return humanizeResult2(r)
	}
}

func humanizeResult2(r resource.TestResult) string {
	if r.Err != nil {
		return red("%s: %s: Error: %s", r.ResourceId, r.Property, r.Err)
	}

	switch r.TestType {
	case resource.Value:
		if r.Successful {
			return green("%s: %s: %s: matches expectation: %s", r.ResourceType, r.ResourceId, r.Property, r.Expected)
		} else {
			return red("%s: %s: %s: doesn't match, expect: %s found: %s", r.ResourceType, r.ResourceId, r.Property, r.Expected, r.Found)
		}
	case resource.Values:
		if r.Successful {
			return green("%s: %s: %s: all expectations found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(r.Expected, ", "))
		} else {
			return red("%s: %s: %s: expectations not found [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(subtractSlice(r.Expected, r.Found), ", "))
		}
	case resource.Contains:
		if r.Successful {
			return green("%s: %s: %s: all patterns found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(r.Expected, ", "))
		} else {
			return red("%s: %s: %s: patterns not found: [%s]", r.ResourceType, r.ResourceId, r.Property, strings.Join(subtractSlice(r.Expected, r.Found), ", "))
		}
	default:
		return red("Unexpected type %d", r.TestType)
	}
}

// Copied from database/sql
var (
	outputersMu sync.Mutex
	outputers   = make(map[string]Outputer)
)

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

func GetOutputer(name string) Outputer {
	if _, ok := outputers[name]; !ok {
		fmt.Println("goss: Bad output format: " + name)
		os.Exit(1)
	}
	return outputers[name]
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
