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
// Accept header, against a passing and a failing suite.
func TestServeStatusMatrix(t *testing.T) {
	failingStatus := map[string]int{
		"documentation": http.StatusServiceUnavailable,
		"json":          http.StatusServiceUnavailable,
		"junit":         http.StatusServiceUnavailable,
		"nagios":        http.StatusServiceUnavailable,
		"prometheus":    http.StatusServiceUnavailable,
		"rspecish":      http.StatusServiceUnavailable,
		"silent":        http.StatusServiceUnavailable,
		"structured":    http.StatusServiceUnavailable,
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
				want := http.StatusOK
				if suite == "failing" {
					want = failingStatus[format]
				}

				logger, _ := captureRecords(slog.LevelDebug)
				config, err := util.NewConfig(
					util.WithSpecFile(specFile),
					// Content-Type also proves that the negotiated format was used
					// rather than silently falling back to this configured format.
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
				contentType := "application/vnd.goss-" + format
				switch format {
				case "json":
					contentType = "application/json"
				case "prometheus":
					contentType = "text/plain; version=0.0.4"
				}
				require.Equal(t, contentType, rr.Result().Header.Get("Content-Type"))
			})
		}
	}
}
