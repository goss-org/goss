package outputs

import (
	"io"
	"strings"
	"time"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/expfmt"
)

const (
	labelType       = "type"
	labelOutcome    = "outcome"
	labelResourceId = "resource_id"
)

var (
	registry      *prometheus.Registry
	testOutcomes  *prometheus.CounterVec
	testDurations *prometheus.CounterVec
	runOutcomes   *prometheus.CounterVec
	runDuration   *prometheus.CounterVec
)

// Prometheus renders metrics in prometheus.io text-format https://prometheus.io/docs/instrumenting/exposition_formats/#text-based-format
type Prometheus struct{}

// ValidOptions is a list of valid format options for prometheus
func (r Prometheus) ValidOptions() []*formatOption {
	return []*formatOption{
		{name: foVerbose},
	}
}

// Output converts the results into the prometheus text-format.
func (r Prometheus) Output(w io.Writer, results <-chan []resource.TestResult,
	outConfig util.OutputConfig) (exitCode int) {
	verbose := util.IsValueInList(foVerbose, outConfig.FormatOptions)

	if registry == nil {
		setupMetrics(verbose)
	}

	overallOutcome := resource.OutcomeUnknown
	var startTime time.Time
	for resultGroup := range results {
		for i, tr := range resultGroup {
			if startTime.IsZero() || tr.StartTime.Before(startTime) {
				startTime = tr.StartTime
			}
			resType := strings.ToLower(tr.ResourceType)
			outcome := tr.ToOutcome()
			if verbose {
				resId := tr.ResourceId
				testOutcomes.WithLabelValues(resType, outcome, resId).Inc()
				testDurations.WithLabelValues(resType, outcome, resId).Add(float64(tr.Duration.Milliseconds()))
			} else {
				testOutcomes.WithLabelValues(resType, outcome).Inc()
				testDurations.WithLabelValues(resType, outcome).Add(float64(tr.Duration.Milliseconds()))
			}
			if i == 0 || canChangeOverallOutcome(overallOutcome, outcome) {
				overallOutcome = outcome
			}
		}
	}

	runOutcomes.WithLabelValues(overallOutcome).Inc()
	runDuration.WithLabelValues(overallOutcome).Add(float64(time.Since(startTime).Milliseconds()))

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

func setupMetrics(verbose bool) {
	registry = prometheus.NewRegistry()
	factory := promauto.With(registry)

	var testLabels []string
	if verbose {
		testLabels = []string{labelType, labelOutcome, labelResourceId}
	} else {
		testLabels = []string{labelType, labelOutcome}
	}

	testOutcomes = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "outcomes_total",
		Help:      "The number of test-outcomes from this run.",
	}, testLabels)
	testDurations = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "outcomes_duration_milliseconds",
		Help:      "The duration of tests from this run. Note; tests run concurrently.",
	}, testLabels)
	runOutcomes = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "run_outcomes_total",
		Help:      "The outcomes of this run as a whole.",
	}, []string{labelOutcome})
	runDuration = factory.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "run_duration_milliseconds",
		Help:      "The end-to-end duration of this run.",
	}, []string{labelOutcome})
}

func canChangeOverallOutcome(current, result string) bool {
	switch current {
	case resource.OutcomeSkip:
		return true
	case resource.OutcomeFail:
		return false
	case resource.OutcomePass:
		return result != resource.OutcomeSkip
	default:
		return result == resource.OutcomeFail
	}
}
