package outputs

import (
	"fmt"
	"strings"

	"github.com/aelsabbahy/goss/resource"
	"github.com/codegangsta/cli"
)

type Outputer interface {
	Output(<-chan resource.TestResult, *cli.Context) bool
}

type Rspecish struct{}

func (r Rspecish) Output(results <-chan resource.TestResult, c *cli.Context) (hasFail bool) {
	testCount := 0
	var failed []resource.TestResult
	for testResult := range results {
		//fmt.Printf("%v: %s.\n", testResult.Duration, testResult.Desc)
		if testResult.Result {
			//fmt.Printf(".")
			fmt.Println(rspecishFmtResult(testResult))
			//fmt.Printf("\n%s\n", testResult.Desc)
			testCount++
		} else {
			//fmt.Printf("F")
			fmt.Println(rspecishFmtResult(testResult))
			//fmt.Printf("\n%s\n", testResult.Desc)
			failed = append(failed, testResult)
			testCount++
		}
	}

	if len(failed) > 0 {
		fmt.Println("\n\nFailures:")
		for _, testResult := range failed {
			fmt.Println(rspecishFmtResult(testResult))
			//fmt.Printf("\n%s\n", testResult.Desc)
		}
	}

	fmt.Printf("\n\nCount: %d failed: %d\n", testCount, len(failed))
	if len(failed) > 0 {
		return true
	}
	return false
}

func rspecishFmtResult(r resource.TestResult) string {
	if r.Err != nil {
		return fmt.Sprintf("%s: %s: Error: %s", r.Title, r.Property, r.Err)
	}
	switch r.Type {
	case resource.Values:
		if r.Result {
			return fmt.Sprintf("%s: %s matches", r.Title, r.Property)
		} else {
			return fmt.Sprintf("%s: %s: [%s] not in: [%s]", r.Title, r.Property, strings.Join(r.Expected, ", "), strings.Join(r.Found, ", "))
		}
	case resource.Value:
		if r.Result {
			return fmt.Sprintf("%s: %s matches", r.Title, r.Property)
		} else {
			return fmt.Sprintf("%s: %s: [%s] not in: [%s]", r.Title, r.Property, strings.Join(r.Expected, ", "), strings.Join(r.Found, ", "))
		}
	case resource.Contains:
		if r.Result {
			return fmt.Sprintf("%s: %s matches", r.Title, r.Property)
		} else {
			return fmt.Sprintf("%s: %s: doesn't match, expect: [%s] found: [%s]", r.Title, r.Property, strings.Join(r.Expected, ", "), strings.Join(r.Found, ", "))
		}
	default:
		return fmt.Sprintf("Unexpected type %d", r.Type)
	}
}
