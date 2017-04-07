package outputs

import (
	"fmt"
	"io"
	"time"
  "strings"

	"github.com/aelsabbahy/goss/resource"

)

type Prometheus struct{}

func (r Prometheus) Output(w io.Writer, results <-chan []resource.TestResult, startTime time.Time) (exitCode int) {
	time_now_ms := int64(time.Now().UnixNano()) / 1000000
	testCount := 0
	failed := 0
	success := 0
	skipped := 0

	var summary map[int]string
	summary = make(map[int]string)

	for resultGroup := range results {
		for _, testResult := range resultGroup {
			switch testResult.Result {
			case resource.SUCCESS:
				success++
			case resource.FAIL:
				failed++
			case resource.SKIP:
				skipped++
			default:
				panic(fmt.Sprintf("Unexpected Result Code: %v\n", testResult.Result))
			}
			summary[testCount] = fmt.Sprintf("%s %d\n", mechanizeResult(testResult), time_now_ms)
			testCount++
		}
	}

	for i := 0; i < testCount; i++ {
		fmt.Fprintf(w, "%s", summary[i])
	}

  // Print goss run metrics
	fmt.Fprintf(w, "goss_count %d\n", testCount)
	fmt.Fprintf(w, "goss_success_count %d\n", success)
	fmt.Fprintf(w, "goss_skipped_count %d\n", skipped)
	fmt.Fprintf(w, "goss_failed_count %d\n", failed)

	duration := float64(time.Since(startTime).Seconds())
	fmt.Fprintf(w, "goss_duration_seconds %.3f\n", duration)

	if failed > 0 {
		return 1
	}

	return 0
}

func init() {
	RegisterOutputer("prometheus", &Prometheus{})
}

func mechanizeResult(r resource.TestResult) string {
	resource_name := fmt.Sprintf("goss_%s_duration_ms", strings.ToLower(r.ResourceType))
	return fmt.Sprintf("%s{resource_id=\"%s\",property=\"%s\"} %d",
		resource_name, r.ResourceId, r.Property, int64(r.Duration) / 1000000)
}
