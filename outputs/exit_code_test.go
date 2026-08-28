package outputs

import (
	"io"
	"log/slog"
	"testing"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// TestOutputerExitCodes exists because this change touches two of the nine
// Output methods. Every registered outputer is driven with a passing and a
// failing suite, and must return exactly what it returned before.
// One row is wrong and is pinned anyway: structured returns 0 for a failing
// suite. Fixing that is a behaviour change that has no business hiding inside a
// logging change.
func TestOutputerExitCodes(t *testing.T) {
	tests := map[string]struct {
		passing int
		failing int
	}{
		"documentation": {passing: 0, failing: 1},
		"json":          {passing: 0, failing: 1},
		"junit":         {passing: 0, failing: 1},
		"nagios":        {passing: 0, failing: 2},
		"prometheus":    {passing: 0, failing: 1},
		"rspecish":      {passing: 0, failing: 1},
		"silent":        {passing: 0, failing: 1},
		"structured":    {passing: 0, failing: 0},
		"tap":           {passing: 0, failing: 1},
	}

	require.Len(t, tests, len(Outputers()),
		"every registered outputer needs a row here")

	for name, want := range tests {
		outputer, err := GetOutputer(name)
		require.NoError(t, err)

		rows := map[string]struct {
			results func() <-chan []resource.TestResult
			want    int
		}{
			"passing": {results: passingResults, want: want.passing},
			"failing": {results: failingResults, want: want.failing},
		}

		for suite, row := range rows {
			t.Run(name+"/"+suite, func(t *testing.T) {
				// The prometheus outputer accumulates into a package-level
				// registry, so leaving it populated would change what the next
				// test in this package sees.
				defer resetMetrics()

				// A logger is injected so that the matrix covers the converted
				// code paths rather than only their silent variants.
				logger, _ := captureRecords(slog.LevelDebug)

				got := outputer.Output(io.Discard, row.results(), util.OutputConfig{Logger: logger})
				require.Equal(t, row.want, got)
			})
		}
	}
}
