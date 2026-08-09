package outputs

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
)

// TestOutputersAreConcurrencySafe is the regression test for the data race on
// github.com/fatih/color's package-level NoColor variable.
//
// `goss serve` runs an Outputer once per in-flight health probe, and the
// machine-readable formatters disable colorization on every invocation. Before
// SetNoColor existed those assignments raced against the colorizing helpers
// reading the same variable from another request. Run this under -race.
func TestOutputersAreConcurrencySafe(t *testing.T) {
	outputers := map[string]Outputer{
		"json":          Json{},
		"junit":         JUnit{},
		"rspecish":      Rspecish{},
		"documentation": Documentation{},
		"tap":           Tap{},
		"nagios":        Nagios{},
	}

	for name, outputer := range outputers {
		t.Run(name, func(t *testing.T) {
			const n = 16
			var wg sync.WaitGroup
			wg.Add(n)
			for range n {
				go func() {
					defer wg.Done()
					outputer.Output(io.Discard, mixedResults(), util.OutputConfig{})
				}()
			}
			wg.Wait()
		})
	}
}

// mixedResults returns a closed channel carrying one result of each outcome, so
// that every colorizing branch of humanizeResult is exercised.
func mixedResults() <-chan []resource.TestResult {
	now := time.Now()
	c := make(chan []resource.TestResult, 1)
	c <- []resource.TestResult{
		{
			Successful: true, Result: resource.SUCCESS, ResourceType: "File",
			ResourceId: "/tmp", Property: "exists",
			StartTime: now, EndTime: now,
		},
		{
			Successful: false, Result: resource.FAIL, ResourceType: "File",
			ResourceId: "/nope", Property: "exists",
			StartTime: now, EndTime: now,
		},
		{
			Successful: true, Result: resource.SKIP, ResourceType: "File",
			ResourceId: "/skip", Property: "exists", Skipped: true,
			StartTime: now, EndTime: now,
		},
	}
	close(c)
	return c
}
