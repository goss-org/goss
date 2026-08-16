package goss

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/goss-org/goss/outputs"
	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// TestServeStatusMatrix covers every registered format, negotiated over the
// Accept header, against a passing and a failing suite. The status is derived
// from the exit code, so structured, the one outputer that returns 0 for a
// failing suite, answers 200 for one. That is pinned here for the same reason
// the exit codes themselves are: it is existing behaviour, and this change is
// not the place to alter it.
func TestServeStatusMatrix(t *testing.T) {
	t.Parallel()

	failingStatus := map[string]int{
		"documentation": http.StatusServiceUnavailable,
		"json":          http.StatusServiceUnavailable,
		"junit":         http.StatusServiceUnavailable,
		"nagios":        http.StatusServiceUnavailable,
		"prometheus":    http.StatusServiceUnavailable,
		"rspecish":      http.StatusServiceUnavailable,
		"silent":        http.StatusServiceUnavailable,
		"structured":    http.StatusOK,
		"tap":           http.StatusServiceUnavailable,
	}

	require.Len(t, failingStatus, len(outputs.Outputers()),
		"every registered outputer needs a row here")

	suites := map[string]string{
		"passing": filepath.Join("testdata", "passing.goss.yaml"),
		"failing": filepath.Join("testdata", "failing.goss.yaml"),
	}

	for format := range failingStatus {
		for suite, specFile := range suites {
			t.Run(format+"/"+suite, func(t *testing.T) {
				t.Parallel()

				want := http.StatusOK
				if suite == "failing" {
					want = failingStatus[format]
				}

				logger, _ := captureRecords(slog.LevelDebug)
				config, err := util.NewConfig(
					util.WithSpecFile(specFile),
					// Deliberately different from the negotiated format, so a
					// row that silently fell back would show up as the wrong
					// status.
					util.WithOutputFormat("silent"),
					util.WithLogger(logger),
				)
				require.NoError(t, err)

				handler, err := newHealthHandler(config)
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
				req.Header.Set("Accept", "application/vnd.goss-"+format)
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, req)

				require.Equal(t, want, rr.Code)
			})
		}
	}
}
