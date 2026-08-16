package goss

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// TestServeRequestRecord covers the per-request record. The status line the old
// one interpolated becomes attributes, and the response body keeps its gate: it
// is attached only when the probe failed.
func TestServeRequestRecord(t *testing.T) {
	tests := map[string]struct {
		specFile   string
		wantStatus int
		wantBody   bool
	}{
		"passing suite": {
			specFile:   filepath.Join("testdata", "passing.goss.yaml"),
			wantStatus: http.StatusOK,
			wantBody:   false,
		},
		"failing suite": {
			specFile:   filepath.Join("testdata", "failing.goss.yaml"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, records := captureRecords(slog.LevelDebug)

			config, err := util.NewConfig(
				util.WithSpecFile(tc.specFile),
				util.WithOutputFormat("json"),
				util.WithLogger(logger),
			)
			require.NoError(t, err)

			handler, err := newHealthHandler(config)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
			require.Equal(t, tc.wantStatus, rr.Code)

			complete := recordsWithMessage(records(), "request complete")
			require.Len(t, complete, 1, "exactly one record per request")
			require.Equal(t, "DEBUG", complete[0]["level"])
			require.Equal(t, float64(tc.wantStatus), complete[0]["http_status"])
			require.Equal(t, "192.0.2.1:1234", complete[0]["client_addr"],
				"httptest's synthetic remote address")

			body, present := complete[0]["response_body"]
			if !tc.wantBody {
				require.False(t, present, "a successful probe must not log its body")
				return
			}

			require.True(t, present, "a failing probe should log its body")
			require.Equal(t, rr.Body.String(), body,
				"the body should be the raw response, with no prose prefix")
		})
	}
}

// TestServeCacheRecords replaces the regression test for
// https://github.com/goss-org/goss/issues/991. The cache-miss message flooded
// the logs of any container behind a load balancer because it carried no level
// prefix for the filter to match. Under slog the level is a property of the
// record, so the same guarantee is now structural rather than textual.
func TestServeCacheRecords(t *testing.T) {
	const cacheTTL = 100 * time.Millisecond

	tests := map[string]struct {
		level   slog.Level
		visible bool
	}{
		"trace shows it": {level: util.LevelTrace, visible: true},
		"debug mutes it": {level: slog.LevelDebug, visible: false},
		"info mutes it":  {level: slog.LevelInfo, visible: false},
		"error mutes it": {level: slog.LevelError, visible: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			logger, records := captureRecords(tc.level)

			config, err := util.NewConfig(
				util.WithSpecFile(filepath.Join("testdata", "matching_basic.yaml")),
				util.WithOutputFormat("json"),
				util.WithCache(cacheTTL),
				util.WithLogger(logger),
			)
			require.NoError(t, err)

			handler, err := newHealthHandler(config)
			require.NoError(t, err)

			probe := func() {
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
				require.Equal(t, http.StatusOK, rr.Code)
			}

			probe() // cold cache
			probe() // warm cache
			time.Sleep(cacheTTL + 5*time.Millisecond)
			probe() // expired cache

			misses := recordsWithMessage(records(), "running validation for stale cache")
			hits := recordsWithMessage(records(), "returning cached result")

			if !tc.visible {
				require.Empty(t, misses, "cache misses should be muted at %s", tc.level)
				require.Empty(t, hits)
				return
			}

			require.Len(t, misses, 2, "cold and expired")
			require.Len(t, hits, 1, "warm")
			for _, record := range append(misses, hits...) {
				require.Equal(t, "TRACE", record["level"])
				require.Equal(t, "res", record["cache_key"])
			}
		})
	}
}

// TestServeNegotiationFallbackRecord covers the one remaining ServeHTTP site,
// where the error becomes an attribute rather than being formatted into the
// message.
func TestServeNegotiationFallbackRecord(t *testing.T) {
	logger, records := captureRecords(slog.LevelDebug)

	config, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "matching_basic.yaml")),
		util.WithOutputFormat("json"),
		util.WithLogger(logger),
	)
	require.NoError(t, err)

	handler, err := newHealthHandler(config)
	require.NoError(t, err)

	// No Accept header at all, which is what makes negotiation fail.
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	fallback := recordsWithMessage(records(), "using configured output format")
	require.Len(t, fallback, 1)
	require.Equal(t, "DEBUG", fallback[0]["level"])
	require.Contains(t, fallback[0]["error"], "Accept header")
}

// TestHTTPServerErrorsReachTheLogger drives the ErrorLog bridge. A handler
// panic is the one stimulus that reliably reaches http.Server.ErrorLog; accept
// and TLS handshake failures need conditions a unit test cannot arrange.
func TestHTTPServerErrorsReachTheLogger(t *testing.T) {
	logger, records := captureRecords(slog.LevelError)

	config, err := util.NewConfig(util.WithLogger(logger))
	require.NoError(t, err)

	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("deliberate panic from a test handler")
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := newServer(config, panicking)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })

	// The panic kills the connection, so the client always errors here.
	//nolint:bodyclose // there is no response to close
	_, err = http.Get("http://" + listener.Addr().String())
	require.Error(t, err)

	var errorRecords []map[string]any
	require.Eventually(t, func() bool {
		errorRecords = records()
		return len(errorRecords) > 0
	}, 5*time.Second, 10*time.Millisecond, "the server error should reach the injected logger")

	require.Equal(t, "ERROR", errorRecords[0]["level"])
	require.Contains(t, errorRecords[0]["msg"], "panic",
		"the standard library's message is passed through opaquely")
}

// TestServerErrorLogIsSilentWithoutALogger keeps the bridge silent without a
// logger: it is built from the guarded logger, so no logger means no records
// anywhere.
func TestServerErrorLogIsSilentWithoutALogger(t *testing.T) {
	config, err := util.NewConfig()
	require.NoError(t, err)
	require.Nil(t, config.Logger)

	global := captureGlobalLogger(t)

	server := newServer(config, nil)
	require.NotNil(t, server.ErrorLog)
	server.ErrorLog.Print("a server error nobody asked to see")

	require.Empty(t, global())
}

// TestNewServerPreservesDefaultServeMux pins the deliberate nil handler: goss
// still serves from http.DefaultServeMux, and swapping that out is a separate
// change.
func TestNewServerPreservesDefaultServeMux(t *testing.T) {
	config, err := util.NewConfig()
	require.NoError(t, err)
	config.ListenAddress = "127.0.0.1:65000"

	server := newServer(config, nil)

	require.Nil(t, server.Handler, "a nil handler is what selects DefaultServeMux")
	require.Equal(t, "127.0.0.1:65000", server.Addr)
}
