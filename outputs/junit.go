package outputs

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/aelsabbahy/goss/resource"
)

type JUnit struct{}

func (r JUnit) Output(w io.Writer, results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	var testCount, failed, skipped int

	// ISO8601 timeformat
	timestamp := time.Now().Format(time.RFC3339)

	var summary map[int]string
	summary = make(map[int]string)

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			m := struct2map(testResult)
			duration := strconv.FormatFloat(m["duration"].(float64)/1000/1000/1000, 'f', 3, 64)
			summary[testCount] = "<testcase name=\"" +
				testResult.ResourceType + " " +
				escapeString(testResult.ResourceId) + " " +
				testResult.Property + "\" " +
				"time=\"" + duration + "\">\n"
			if testResult.Result == resource.FAIL {
				summary[testCount] += "<system-err>" +
					escapeString(humanizeResult2(testResult)) +
					"</system-err>\n"
				summary[testCount] += "<failure>" +
					escapeString(humanizeResult2(testResult)) +
					"</failure>\n</testcase>\n"

				failed++
			} else {
				if testResult.Result == resource.SKIP {
					summary[testCount] += "<skipped/>"
					skipped++
				}
				summary[testCount] += "<system-out>" +
					escapeString(humanizeResult2(testResult)) +
					"</system-out>\n</testcase>\n"
			}
			testCount++
		}
	}

	duration := time.Since(startTime)
	fmt.Fprintln(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	fmt.Fprintf(w, "<testsuite name=\"goss\" errors=\"0\" tests=\"%d\" "+
		"failures=\"%d\" skipped=\"%d\" time=\"%.3f\" timestamp=\"%s\">\n",
		testCount, failed, skipped, duration.Seconds(), timestamp)

	for i := 0; i < testCount; i++ {
		fmt.Fprintf(w, "%s", summary[i])
	}

	fmt.Fprintln(w, "</testsuite>")

	if failed > 0 {
		return 1
	}

	return 0
}

func init() {
	RegisterOutputer("junit", &JUnit{})
}

func escapeString(str string) string {
	buffer := new(bytes.Buffer)
	xml.EscapeText(buffer, []byte(str))
	return buffer.String()
}
