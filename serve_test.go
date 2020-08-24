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
		specFile           string
		expectedHTTPStatus int
	}{
		"passing": {
			specFile:           filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus: 200,
		},
		"failing": {
			specFile:           filepath.Join("testdata", "failing.goss.yaml"),
			expectedHTTPStatus: 503,
		},
	}
	for testName := range tests {
		tc := tests[testName]
		t.Run(testName, func(t *testing.T) {
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)

			config, err := util.NewConfig(util.WithSpecFile(tc.specFile))
			require.NoError(t, err)

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
		})
	}
}
