package outputs

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
	"github.com/fatih/color"
)

type Prometheus struct{}

func (r Prometheus) ValidOptions() []*formatOption {
	return []*formatOption{}
}

func (r Prometheus) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {

	color.NoColor = true
	testCount := 0
	var skipped, failed int

	const resultName string = "goss_result"
	const durationName string = "goss_duration_seconds"
	const promType string = "gauge"

	fmt.Fprintf(w, "# HELP %s goss result\n", resultName)
	fmt.Fprintf(w, "# TYPE %s %s\n", resultName, promType)
	fmt.Fprintf(w, "# HELP %s goss duration\n", durationName)
	fmt.Fprintf(w, "# TYPE %s %s\n", durationName, promType)

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			labels := make(map[string]string)
			labels["property"] = testResult.Property
			labels["resource_id"] = testResult.ResourceId
			labels["resource_type"] = testResult.ResourceType
			switch testResult.Result {
			case resource.SUCCESS:
				labels["result"] = "success"
			case resource.SKIP:
				labels["result"] = "skipped"
				skipped++
			case resource.FAIL:
				labels["result"] = "failure"
				failed++
			}
			testCount++

			if testResult.Title != "" {
				labels["title"] = testResult.Title
			}
			if len(testResult.Meta) > 0 {
				for k, v := range testResult.Meta {
					labels[ToSnakeCase(k)] = v.(string)
				}
			}

			fmt.Fprintf(w, "%s{%s} %d\n", resultName, KeysString(labels), testResult.Result)
			fmt.Fprintf(w, "%s{%s} %f\n", durationName, KeysString(labels), float64(testResult.Duration.Nanoseconds())/1000000000)
		}
	}

	fmt.Fprint(w, "\n\n")
	fmt.Fprint(w, "# HELP goss_failed_total goss failed tests\n")
	fmt.Fprint(w, "# TYPE goss_failed_total gauge\n")
	fmt.Fprintf(w, "goss_failed_total %d\n", failed)
	fmt.Fprint(w, "# HELP goss_successful_total goss successful tests\n")
	fmt.Fprint(w, "# TYPE goss_successful_total gauge\n")
	fmt.Fprintf(w, "goss_successful_total %d\n", testCount-failed-skipped)
	fmt.Fprint(w, "# HELP goss_skipped_total goss skipped tests\n")
	fmt.Fprint(w, "# TYPE goss_skipped_total gauge\n")
	fmt.Fprintf(w, "goss_skipped_total %d\n", skipped)
	fmt.Fprint(w, "# HELP goss_test_count goss test count\n")
	fmt.Fprint(w, "# TYPE goss_test_count gauge\n")
	fmt.Fprintf(w, "goss_test_count %d\n", testCount)
	fmt.Fprint(w, "# HELP goss_duration_seconds_total goss total duration\n")
	fmt.Fprint(w, "# TYPE goss_duration_seconds_total gauge\n")
	fmt.Fprintf(w, "goss_duration_seconds_total %f\n", float64(time.Since(startTime).Nanoseconds())/1000000000)

	return 0
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func QuoteValue(s string) string {
	var vs string = s
	vs = strings.ReplaceAll(vs, "\\", "\\\\")
	vs = strings.ReplaceAll(vs, "\n", "\\n")
	vs = strings.ReplaceAll(vs, "\"", "\\\"")
	return vs
}

func KeysString(m map[string]string) string {
	l := make([]string, 0, len(m))
	for k, v := range m {
		l = append(l, fmt.Sprintf("%s=\"%s\"", k, QuoteValue(v)))
	}
	return strings.Join(l, ",")
}
