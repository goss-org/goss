package goss

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/aelsabbahy/goss/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServe(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		outputFormat        string
		specFile            string
		expectedHTTPStatus  int
		expectedContentType string
	}{
		"passing-json": {
			outputFormat:        "json",
			specFile:            filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus:  http.StatusOK,
			expectedContentType: "application/json",
		},
		"failing-json": {
			outputFormat:        "json",
			specFile:            filepath.Join("testdata", "failing.goss.yaml"),
			expectedHTTPStatus:  http.StatusServiceUnavailable,
			expectedContentType: "application/json",
		},
		"failing-default-output": {
			outputFormat:        "structured",
			specFile:            filepath.Join("testdata", "failing.goss.yaml"),
			expectedHTTPStatus:  http.StatusServiceUnavailable,
			expectedContentType: "",
		},
	}
	for testName := range tests {
		tc := tests[testName]
		t.Run(testName, func(t *testing.T) {
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)

			config, err := util.NewConfig(util.WithSpecFile(tc.specFile), util.WithOutputFormat(tc.outputFormat))
			require.NoError(t, err)
			t.Logf("Config: %v", config)

			hh, err := newHealthHandler(config)
			require.NoError(t, err)

			req, err := http.NewRequest("GET", config.Endpoint, nil)
			if err != nil {
				require.NoError(t, err)
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(hh.ServeHTTP)

			handler.ServeHTTP(rr, req)

			t.Logf("testName %q log output:\n%s", testName, logOutput.String())
			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			if tc.expectedContentType != "" {
				assert.Equal(t, []string{tc.expectedContentType}, rr.HeaderMap["Content-Type"])
			}
		})
	}
}
