package goss

import (
	"path/filepath"
	"testing"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasAnyMark(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		resource []string
		filter   []string
		want     bool
	}{
		{"both empty", nil, nil, false},
		{"resource empty filter set", nil, []string{"a"}, false},
		{"resource set filter empty", []string{"a"}, nil, false},
		{"single match", []string{"a"}, []string{"a"}, true},
		{"no overlap", []string{"a"}, []string{"b"}, false},
		{"partial overlap", []string{"a", "b"}, []string{"b", "c"}, true},
		{"full overlap", []string{"a", "b"}, []string{"a", "b"}, true},
		{"order independent", []string{"b", "a"}, []string{"c", "a"}, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := hasAnyMark(tc.resource, tc.filter); got != tc.want {
				t.Errorf("hasAnyMark(%v, %v) = %v, want %v", tc.resource, tc.filter, got, tc.want)
			}
		})
	}
}

func TestShouldRunByMarks(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		resource []string
		include  []string
		exclude  []string
		want     bool
	}{
		// No filters set - everything runs
		{"no filters, no resource marks", nil, nil, nil, true},
		{"no filters, with resource marks", []string{"critical"}, nil, nil, true},

		// Include filter set
		{"include set, resource matches", []string{"critical"}, []string{"critical"}, nil, true},
		{"include set, resource matches one of many", []string{"critical", "fast"}, []string{"critical"}, nil, true},
		{"include set, resource does not match", []string{"slow"}, []string{"critical"}, nil, false},
		{"include set, resource has no marks", nil, []string{"critical"}, nil, false},
		{"include set, multiple include marks, partial match", []string{"network"}, []string{"critical", "network"}, nil, true},

		// Exclude filter set
		{"exclude set, resource matches exclusion", []string{"slow"}, nil, []string{"slow"}, false},
		{"exclude set, resource does not match", []string{"fast"}, nil, []string{"slow"}, true},
		{"exclude set, resource has no marks", nil, nil, []string{"slow"}, true},
		{"exclude set, multiple exclude marks, partial match", []string{"flaky"}, nil, []string{"slow", "flaky"}, false},

		// Both filters set
		{"both set, included and not excluded", []string{"critical", "fast"}, []string{"critical"}, []string{"slow"}, true},
		{"both set, included but excluded", []string{"critical", "slow"}, []string{"critical"}, []string{"slow"}, false},
		{"both set, not included", []string{"network"}, []string{"critical"}, []string{"slow"}, false},
		{"both set, not included nor excluded", []string{"random"}, []string{"critical"}, []string{"slow"}, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldRunByMarks(tc.resource, tc.include, tc.exclude); got != tc.want {
				t.Errorf("shouldRunByMarks(%v, include=%v, exclude=%v) = %v, want %v",
					tc.resource, tc.include, tc.exclude, got, tc.want)
			}
		})
	}
}

// collectResults drains a TestResult channel and classifies outcomes. Used by
// the end-to-end marks tests below to assert on how many resources were
// executed vs. skipped under various filter combinations.
func collectResults(t *testing.T, ch <-chan []resource.TestResult) (ran, skipped int) {
	t.Helper()
	for batch := range ch {
		for _, r := range batch {
			if r.Result == resource.SKIP {
				skipped++
			} else {
				ran++
			}
		}
	}
	return ran, skipped
}

// TestValidateResults_WithIncludeMarks verifies the full
// gossfile -> Config -> ValidateResults -> []TestResult path honors
// IncludeMarks. This complements the unit tests on shouldRunByMarks by
// exercising the same integration path the CLI uses.
//
// testdata/marks.goss.yaml contains three commands with marks:
//   - "echo critical" marked [critical, fast]  -> 2 properties (exit-status, stdout)
//   - "echo slow"     marked [slow]            -> 2 properties
//   - "echo unmarked" no marks                 -> 2 properties
//
// exit-status and stdout are 2 properties each; stderr is empty list so
// no property emitted for it. Total: 6 TestResults per run.
func TestValidateResults_WithIncludeMarks(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithIncludeMarks("critical"),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	results, err := ValidateResults(cfg)
	require.NoError(t, err)

	ran, skipped := collectResults(t, results)
	// 2 properties for the critical command, 4 for the filtered-out ones.
	assert.Equal(t, 2, ran, "only critical command properties should run")
	assert.Equal(t, 4, skipped, "slow and unmarked command properties should skip")
}

// TestValidateResults_WithExcludeMarks verifies ExcludeMarks filters tests.
func TestValidateResults_WithExcludeMarks(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithExcludeMarks("slow"),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	results, err := ValidateResults(cfg)
	require.NoError(t, err)

	ran, skipped := collectResults(t, results)
	// Critical + unmarked run (4 props), slow skipped (2 props).
	assert.Equal(t, 4, ran, "critical and unmarked should run, slow should skip")
	assert.Equal(t, 2, skipped, "only the slow command's properties should skip")
}

// TestValidateResults_IncludeThenExcludeOrder verifies the include-first-then-
// exclude evaluation order. AC-E4 / FR-7.
func TestValidateResults_IncludeThenExcludeOrder(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithIncludeMarks("critical", "slow"),
		util.WithExcludeMarks("slow"),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	results, err := ValidateResults(cfg)
	require.NoError(t, err)

	ran, skipped := collectResults(t, results)
	// Include {critical, slow} then exclude {slow} => only critical runs.
	assert.Equal(t, 2, ran, "only critical should survive include-then-exclude")
	assert.Equal(t, 4, skipped, "slow (excluded) and unmarked (not included) should skip")
}

// TestValidateResults_NoFilters verifies the baseline: with no mark filters,
// all tests run. Regression guard for backward compatibility (NFR-2).
func TestValidateResults_NoFilters(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	results, err := ValidateResults(cfg)
	require.NoError(t, err)

	ran, skipped := collectResults(t, results)
	assert.Equal(t, 6, ran, "with no mark filters, all properties should run")
	assert.Equal(t, 0, skipped, "nothing should be skipped")
}

// TestValidateResults_AllFilteredOut verifies that when a mark filter matches
// no resources, the validator still completes cleanly with every test
// reported as skipped (AC-UW3).
func TestValidateResults_AllFilteredOut(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithIncludeMarks("nonexistent-mark"),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	results, err := ValidateResults(cfg)
	require.NoError(t, err)

	ran, skipped := collectResults(t, results)
	assert.Equal(t, 0, ran, "no tests should run when include filter matches nothing")
	assert.Equal(t, 6, skipped, "all 6 properties should be skipped")
}

// TestValidate_LogsMarkFilterSummary verifies that when a mark filter is
// active, validate() emits a [DEBUG] summary log describing the filter and
// a post-run count of filtered resources. When no filter is set, the logger
// must remain silent on this topic (regression guard for backward compat).
func TestValidate_LogsMarkFilterSummary(t *testing.T) {
	t.Parallel()

	t.Run("filter active emits debug summary and count", func(t *testing.T) {
		t.Parallel()
		tl := util.NewTestLogger(t)

		cfg, err := util.NewConfig(
			util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
			util.WithIncludeMarks("critical"),
			util.WithLogger(tl),
		)
		require.NoError(t, err)

		results, err := ValidateResults(cfg)
		require.NoError(t, err)
		// Drain results so the filtering goroutine runs to completion and
		// emits its post-loop count line.
		_, _ = collectResults(t, results)

		out := tl.String()
		assert.Contains(t, out, "[DEBUG] mark filters active",
			"should announce which filters are in play")
		assert.Contains(t, out, "include=[critical]")
		assert.Contains(t, out, "[DEBUG] marks filter skipped",
			"should summarize filtered count after the run")
	})

	t.Run("no filter means no mark log output", func(t *testing.T) {
		t.Parallel()
		tl := util.NewTestLogger(t)

		cfg, err := util.NewConfig(
			util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
			util.WithLogger(tl),
		)
		require.NoError(t, err)

		results, err := ValidateResults(cfg)
		require.NoError(t, err)
		_, _ = collectResults(t, results)

		out := tl.String()
		assert.NotContains(t, out, "mark filters active",
			"must not log filter-active summary when no filters are set")
		assert.NotContains(t, out, "marks filter skipped",
			"must not log filter-count summary when no filters are set")
	})
}
