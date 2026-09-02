package resource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/goss-org/goss/system"
)

type Resource interface {
	Validate(ctx context.Context, sys *system.System) []TestResult
	SetID(string)
	SetSkip()
	TypeKey() string
	TypeName() string
}

var (
	resourcesMu sync.Mutex
	resources   = map[string]Resource{}
)

func registerResource(key string, resource Resource) {
	resourcesMu.Lock()
	resources[key] = resource
	resourcesMu.Unlock()
}

func Resources() map[string]Resource {
	return resources
}

type ResourceRead interface {
	ID() string
	GetTitle() string
	GetMeta() meta
}

type Retryable interface {
	GetRetryCount() int
	GetRetryDelay() RetryDelay
}

type matcher any
type meta map[string]any

func contains(a []string, s string) bool {
	for _, e := range a {
		if m, _ := filepath.Match(e, s); m {
			return true
		}
	}
	return false
}

func deprecateAtoI(depr any, desc string) any {
	s, ok := depr.(string)
	if !ok {
		return depr
	}
	fmt.Fprintf(os.Stderr, "DEPRECATION WARNING: %s should be an integer not a string\n", desc)
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return float64(i)
}

func shouldSkip(results []TestResult) bool {
	if len(results) < 1 {
		return false
	}
	if results[0].Err != nil || results[0].Result != SUCCESS || results[0].MatcherResult.Actual == false {
		return true
	}
	return false
}

func isSet(i interface{}) bool {
	switch v := i.(type) {
	case []interface{}:
		return len(v) > 0
	default:
		return i != nil
	}
}

// isSetWarnEmpty reports whether a matcher property holds at least one
// condition, warning first when it holds a list written out as empty.
//
// A list of patterns is a list of conditions that must all hold, so an empty
// list is zero conditions: the property is skipped and always passes, which is
// rarely what someone writing it out by hand intends. The behaviour is
// deliberately left alone — goss add file emits contents: [] to record that it
// captured no expectation, and erroring would invalidate every gossfile it has
// ever generated. desc follows the same "<id>: <type>.<property>" shape as the
// deprecation warnings.
func isSetWarnEmpty(i any, desc string) bool {
	if v, ok := i.([]any); ok && len(v) == 0 {
		fmt.Fprintf(os.Stderr, "WARNING: %s is an empty list, which asserts nothing and always passes. Use \"\" to assert empty content, or remove it.\n", desc)
	}
	return isSet(i)
}
