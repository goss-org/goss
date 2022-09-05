package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/aelsabbahy/goss/system"
	"github.com/oleiade/reflections"
)

type Resource interface {
	Validate(*system.System) []TestResult
	SetID(string)
}

var (
	resourcesMu sync.Mutex
	resources   = map[string]Resource{
		"addr":         &Addr{},
		"command":      &Command{},
		"dns":          &DNS{},
		"file":         &File{},
		"gossfile":     &Gossfile{},
		"group":        &Group{},
		"http":         &HTTP{},
		"interface":    &Interface{},
		"kernel-param": &KernelParam{},
		"mount":        &Mount{},
		"package":      &Package{},
		"port":         &Port{},
		"process":      &Process{},
		"service":      &Service{},
		"user":         &User{},
	}
)

func Resources() map[string]Resource {
	return resources
}

type ResourceRead interface {
	ID() string
	GetTitle() string
	GetMeta() meta
}

type matcher interface{}
type meta map[string]interface{}

func contains(a []string, s string) bool {
	for _, e := range a {
		if m, _ := filepath.Match(e, s); m {
			return true
		}
	}
	return false
}

func deprecateAtoI(depr interface{}, desc string) interface{} {
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

func validAttrs(i interface{}, t string) (map[string]bool, error) {
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
