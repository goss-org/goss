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
				summary[testCount] = mechanizeResult(testResult) + "\n"
				success++
			case resource.FAIL:
				summary[testCount] = mechanizeResult(testResult) + "\n"
				failed++
			case resource.SKIP:
				summary[testCount] = mechanizeResult(testResult) + "\n"
				skipped++
			default:
				panic(fmt.Sprintf("Unexpected Result Code: %v\n", testResult.Result))
			}
			testCount++
		}
	}

	for i := 0; i < testCount; i++ {
		fmt.Fprintf(w, "%s", summary[i])
	}

  // Print goss run metrics
	fmt.Fprintf(w, "goss_count %d \n", testCount)
	fmt.Fprintf(w, "goss_success_count %d \n", success)
	fmt.Fprintf(w, "goss_skipped_count %d \n", skipped)
	fmt.Fprintf(w, "goss_failed_count %d \n", failed)

	duration := float64(time.Since(startTime).Seconds())
	fmt.Fprintf(w, "goss_duration_seconds %.3f \n", duration)

	if failed > 0 {
		return 1
	}

	return 0
}

func init() {
	RegisterOutputer("prometheus", &Prometheus{})
}

func mechanizeResult(r resource.TestResult) string {
	var res string
  var err string

	if r.Err != nil {
		err = fmt.Sprintf(",error=\"%v\"", r.Err)
		res = "error"
	} else {
	  switch r.Result {
	  case resource.SUCCESS:
      res = "success"
 	  case resource.SKIP:
      res = "skipped"
	  case resource.FAIL:
	  	res = "failed"
	  default:
	  	panic(fmt.Sprintf("Unexpected Result Code: %v\n", r.Result))
	  }
	}
	return fmt.Sprintf("goss_%s{resource_id=\"%s\",property=\"%s\",result=\"%s\"%s} %d", strings.ToLower(r.ResourceType), r.ResourceId, r.Property, res, err, int64(r.Duration))
}
