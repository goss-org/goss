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
	labelType    = "type"
	labelOutcome = "outcome"
)

var (
	testOutcomes = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "outcomes_total",
		Help:      "The number of test-outcomes from this run.",
	}, []string{labelType, labelOutcome})
	testDurations = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "outcomes_duration_milliseconds",
		Help:      "The duration of tests from this run. Note; tests run concurrently.",
	}, []string{labelType, labelOutcome})
	runOutcomes = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "run_outcomes_total",
		Help:      "The outcomes of this run as a whole.",
	}, []string{labelOutcome})
	runDuration = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goss",
		Subsystem: "tests",
		Name:      "run_duration_milliseconds",
		Help:      "The end-to-end duration of this run.",
	}, []string{labelOutcome})
)

// Prometheus renders metrics in prometheus.io text-format https://prometheus.io/docs/instrumenting/exposition_formats/#text-based-format
type Prometheus struct{}

// NewPrometheus creates and initialises a new Prometheus Outputer (to avoid missing metrics)
func NewPrometheus() *Prometheus {
	outputer := &Prometheus{}
	outputer.init()
	return outputer
}

func (r *Prometheus) init() {
	// Avoid missing metrics: https://prometheus.io/docs/practices/instrumentation/#avoid-missing-metrics
	for resourceType := range resource.Resources() {
		for _, outcome := range resource.HumanOutcomes() {
			testOutcomes.WithLabelValues(resourceType, outcome).Add(0)
			testDurations.WithLabelValues(resourceType, outcome).Add(0)
		}
	}
	runOutcomes.WithLabelValues(labelOutcome).Add(0)
	runDuration.WithLabelValues(labelOutcome).Add(0)
}

// ValidOptions is a list of valid format options for prometheus
func (r Prometheus) ValidOptions() []*formatOption {
	return []*formatOption{}
}

// Output converts the results into the prometheus text-format.
func (r Prometheus) Output(w io.Writer, results <-chan []resource.TestResult,
	startTime time.Time, outConfig util.OutputConfig) (exitCode int) {
	overallOutcome := resource.OutcomeUnknown
	for resultGroup := range results {
		for _, tr := range resultGroup {
			resType := strings.ToLower(tr.ResourceType)
			outcome := tr.ToOutcome()
			testOutcomes.WithLabelValues(resType, outcome).Inc()
			testDurations.WithLabelValues(resType, outcome).Add(float64(tr.Duration.Milliseconds()))
			if tr.Result != resource.SUCCESS {
				overallOutcome = tr.ToOutcome()
			}
		}
	}

	runOutcomes.WithLabelValues(overallOutcome).Inc()
	runDuration.WithLabelValues(overallOutcome).Add(float64(time.Since(startTime).Milliseconds()))

	metricsFamilies, err := prometheus.DefaultGatherer.Gather()
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
