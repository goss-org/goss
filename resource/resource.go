package resource

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aelsabbahy/goss/system"
	"github.com/oleiade/reflections"
)

type Resource interface {
	Validate(*system.System) []TestResult
	SetID(string)
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
	fmt.Printf("DEPRECATION WARNING: %s should be an integer not a string\n", desc)
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
	if results[0].Err != nil {
		return true
	}
	if len(results[0].Found) < 1 {
		return false
	}
	if results[0].Found == "false" {
		return true
	}
	return false
}
