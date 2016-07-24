package outputs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aelsabbahy/goss/resource"
)

type JUnit struct{}

func (r JUnit) Output(results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	var testCount, failed, skipped int

	// ISO8601 timeformat
	location, _ := time.LoadLocation("Etc/UTC")
	timestamp := time.Now().In(location).Format(time.RFC3339)

	var summary map[int]string
	summary = make(map[int]string)

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			m := struct2map(testResult)
			duration := strconv.FormatFloat(m["duration"].(float64)/1000/1000/1000, 'f', 3, 64)
			summary[testCount] = "<testcase name=\"" +
				testResult.ResourceType + " " +
				testResult.ResourceId + " " +
				testResult.Property + "\" " +
				"time=\"" + duration + "\">\n"
			if testResult.Result == resource.FAIL {
				summary[testCount] += "<system-err>" +
					humanizeResult2(testResult) +
					"</system-err>\n"
				summary[testCount] += "<failure>" +
					humanizeResult2(testResult) +
					"</failure>\n</testcase>\n"

				failed++
			} else {
				if testResult.Result == resource.SKIP {
					summary[testCount] += "<skipped/>"
					skipped++
				}
				summary[testCount] += "<system-out>" +
					humanizeResult2(testResult) +
					"</system-out>\n</testcase>\n"
			}
			testCount++
		}
	}

	duration := time.Since(startTime)
	fmt.Println("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	fmt.Printf("<testsuite name=\"goss\" errors=\"0\" tests=\"%d\" "+
		"failures=\"%d\" skipped=\"%d\" time=\"%.3f\" timestamp=\"%s\">\n",
		testCount, failed, skipped, duration.Seconds(), timestamp)

	for i := 0; i < testCount; i++ {
		fmt.Printf("%s", summary[i])
	}

	fmt.Println("</testsuite>")

	if failed > 0 {
		return 1
	}

	return 0
}

func init() {
	RegisterOutputer("junit", &JUnit{})
}
