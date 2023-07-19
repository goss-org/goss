package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/goss-org/goss/system"
	"github.com/oleiade/reflections"
)

type Resource interface {
	Validate(sys *system.System) []TestResult
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

func validAttrs(i any, t string) (map[string]bool, error) {
	validAttrs := make(map[string]bool)
	tags, err := reflections.Tags(i, t)
	if err != nil {
		return nil, err
	}
	for _, v := range tags {
		validAttrs[strings.Split(v, ",")[0]] = true
	}
	return validAttrs, nil
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
