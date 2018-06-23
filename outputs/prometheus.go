package outputs

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
)

type Prometheus struct{}

func (r Prometheus) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {

	var testCount, success, failed, skipped int
	testCount, success, failed, skipped = 0, 0, 0, 0

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

			summary[testCount] = fmt.Sprintf("%s\n", mechanizeResult(testResult))

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
	RegisterOutputer("prometheus", &Prometheus{}, []string{})
}

func mechanizeResult(r resource.TestResult) string {
	resourceName := fmt.Sprintf("goss_%s", strings.ToLower(r.ResourceType))
	return fmt.Sprintf("%s{resource_id=\"%s\",property=\"%s\"} %d",
		resourceName, r.ResourceId, r.Property, int64(r.Result))
}
