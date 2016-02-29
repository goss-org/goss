package resource

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/aelsabbahy/goss/system"
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
	fmt.Printf("DEPRICATION WARNING: %s should be an integer not a string\n", desc)
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return float64(i)
}
