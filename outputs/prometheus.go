package outputs

import (
	"io"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/expfmt"
)

var (
	registry     = prometheus.NewRegistry()
	testOutcomes = promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "outcomes_total",
		Help:      "The number of test-outcomes from this run.",
	}, []string{"type", "outcome"})
	testDurations = promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "outcomes_duration_seconds",
		Help:      "The duration of tests from this run. Note; tests run concurrently.",
	}, []string{"type", "outcome"})
	runDuration = promauto.With(registry).NewCounter(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "run_duration_seconds",
		Help:      "The end-to-end duration of this run.",
	})
)

// Prometheus renders metrics in prometheus.io text-format https://prometheus.io/docs/instrumenting/exposition_formats/#text-based-format
type Prometheus struct{}

// ValidOptions is a list of valid format options for prometheus
func (r Prometheus) ValidOptions() []*formatOption {
	return []*formatOption{}
}

// Output converts the results into the prometheus text-format.
func (r Prometheus) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {
	for resultGroup := range results {
		for _, tr := range resultGroup {
			resType := strings.ToLower(tr.ResourceType)
			outcome := tr.ToOutcome()
			testOutcomes.WithLabelValues(resType, outcome).Inc()
			testDurations.WithLabelValues(resType, outcome).Add(tr.Duration.Seconds())
		}
	}

	runDuration.Add(time.Since(startTime).Seconds())

	metricsFamilies, err := registry.Gather()
	if err != nil {
		return -1
	}
	encoder := expfmt.NewEncoder(w, expfmt.FmtText)
	for _, mf := range metricsFamilies {
		err := encoder.Encode(mf)
		if err != nil {
			return -1
		}
	}

	return 0
}
