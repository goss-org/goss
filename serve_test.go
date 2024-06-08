package goss

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeWithNoContentNegotiation(t *testing.T) {
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
			outputFormat:        "rspecish",
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

			config, err := util.NewConfig(
				util.WithSpecFile(tc.specFile),
				util.WithOutputFormat(tc.outputFormat),
			)
			require.NoError(t, err)

			hh, err := newHealthHandler(config)
			require.NoError(t, err)

			req := makeRequest(t, config, nil)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(hh.ServeHTTP)

			handler.ServeHTTP(rr, req)

			t.Logf("testName %q log output:\n%s", testName, logOutput.String())
			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			if tc.expectedContentType != "" {
				assert.Equal(t, tc.expectedContentType, rr.Result().Header.Get("Content-Type"))
			}
		})
	}
}

func TestServeNegotiatingContent(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		acceptHeader        []string
		outputFormat        string
		specFile            string
		expectedHTTPStatus  int
		expectedContentType string
	}{
		"accept {blank} returns process-level format-option": {
			acceptHeader: []string{
				"",
			},
			outputFormat:        "structured",
			specFile:            filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus:  http.StatusOK,
			expectedContentType: "application/vnd.goss-structured",
		},
		"accept application/json": {
			acceptHeader: []string{
				"application/json",
			},
			outputFormat:        "structured",
			specFile:            filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus:  http.StatusOK,
			expectedContentType: "application/json",
		},
		"accept text/json translates to application/json": {
			acceptHeader: []string{
				"text/json",
			},
			outputFormat:        "structured",
			specFile:            filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus:  http.StatusOK,
			expectedContentType: "application/json",
		},
		"when accept is application/vnd.goss-json, return more widely known application/json": {
			acceptHeader: []string{
				"application/vnd.goss-json",
			},
			outputFormat:        "structured",
			specFile:            filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus:  http.StatusOK,
			expectedContentType: "application/json",
		},
		"accept header contains vendor-specific output format different from process-level": {
			acceptHeader: []string{
				"application/vnd.goss-rspecish",
			},
			outputFormat:        "structured",
			specFile:            filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus:  http.StatusOK,
			expectedContentType: "application/vnd.goss-rspecish",
		},
		"accept header contains nonsense": {
			acceptHeader: []string{
				"application/vnd.goss-nonexistent",
			},
			outputFormat:        "structured",
			specFile:            filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus:  http.StatusOK,
			expectedContentType: "application/vnd.goss-structured",
		},
		"accept header contains nonsense then valid": {
			acceptHeader: []string{
				"application/vnd.goss-nonexistent",
				"application/json",
			},
			outputFormat:        "structured",
			specFile:            filepath.Join("testdata", "passing.goss.yaml"),
			expectedHTTPStatus:  http.StatusOK,
			expectedContentType: "application/json",
		},
	}
	for testName := range tests {
		tc := tests[testName]
		t.Run(testName, func(t *testing.T) {
			var logOutput bytes.Buffer
			log.SetOutput(&logOutput)

			config, err := util.NewConfig(
				util.WithSpecFile(tc.specFile),
				util.WithOutputFormat(tc.outputFormat),
			)
			require.NoError(t, err)

			hh, err := newHealthHandler(config)
			require.NoError(t, err)

			req := makeRequest(t, config, map[string][]string{
				"accept": tc.acceptHeader,
			})
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(hh.ServeHTTP)

			handler.ServeHTTP(rr, req)

			t.Logf("testName %q log output:\n%s", testName, logOutput.String())
			assert.Equal(t, tc.expectedHTTPStatus, rr.Code)
			if tc.expectedContentType != "" {
				assert.Equal(t, tc.expectedContentType, rr.Result().Header.Get("Content-Type"))
			}
		})
	}
}

func TestServeCacheWithNoContentNegotiation(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	const cache = time.Duration(time.Millisecond * 100)
	config, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "passing.goss.yaml")),
		util.WithCache(cache),
	)
	require.NoError(t, err)

	hh, err := newHealthHandler(config)
	require.NoError(t, err)

	req := makeRequest(t, config, nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(hh.ServeHTTP)

	t.Run("fresh cache", func(t *testing.T) {
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		assert.Contains(t, logOutput.String(), "Stale cache")
		t.Log(logOutput.String())
		logOutput.Reset()
	})

	t.Run("immediately re-request, cache should be warm", func(t *testing.T) {
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		assert.NotContains(t, logOutput.String(), "Stale cache")
		t.Log(logOutput.String())
		logOutput.Reset()
	})

	t.Run("allow cache to expire, cache should be cold", func(t *testing.T) {
		time.Sleep(cache + 5*time.Millisecond)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		assert.Contains(t, logOutput.String(), "Stale cache")
		t.Log(logOutput.String())
		logOutput.Reset()
	})
}

func TestServeCacheNegotiatingContent(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	const cache = time.Duration(time.Millisecond * 100)
	config, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "passing.goss.yaml")),
		util.WithCache(cache),
		util.WithOutputFormat("structured"),
	)
	require.NoError(t, err)

	hh, err := newHealthHandler(config)
	require.NoError(t, err)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(hh.ServeHTTP)

	t.Run("fresh cache", func(t *testing.T) {
		req := makeRequest(t, config, map[string][]string{
			"accept": {"application/json"},
		})
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		assert.Contains(t, logOutput.String(), "Stale cache")
		t.Log(logOutput.String())
		logOutput.Reset()
	})

	t.Run("immediately re-request, cache should be warm", func(t *testing.T) {
		req := makeRequest(t, config, map[string][]string{
			"accept": {"application/json"},
		})
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		assert.NotContains(t, logOutput.String(), "Stale cache")
		t.Log(logOutput.String())
		logOutput.Reset()
	})

	t.Run("immediately re-request but different accept header, cache should be warm", func(t *testing.T) {
		req := makeRequest(t, config, map[string][]string{
			"accept": {"application/vnd.goss-rspecish"},
		})
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		assert.NotContains(t, logOutput.String(), "Stale cache")
		t.Log(logOutput.String())
		logOutput.Reset()
	})

	t.Run("allow cache to expire, cache should be cold", func(t *testing.T) {
		time.Sleep(cache + 5*time.Millisecond)
		req := makeRequest(t, config, map[string][]string{
			"accept": {"application/json"},
		})
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
		assert.Contains(t, logOutput.String(), "Stale cache")
		t.Log(logOutput.String())
		logOutput.Reset()
	})
}

func makeRequest(t *testing.T, config *util.Config, headers map[string][]string) *http.Request {
	req, err := http.NewRequest("GET", config.Endpoint, nil)
	require.NoError(t, err)
	for header, vals := range headers {
		for _, v := range vals {
			req.Header.Add(header, v)
		}
	}
	return req
}
