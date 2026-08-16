package goss

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// TestServeHandlesConcurrentRequests drives a single healthHandler from many
// goroutines, which is what net/http does for every `goss serve` deployment.
// Its value is under -race: it covers the whole per-request path, so a future
// change that reaches for process-wide mutable state from a request will be
// caught here rather than in production.
func TestServeHandlesConcurrentRequests(t *testing.T) {
	t.Parallel()

	// A spec with no subprocesses, so the test stays fast when every goroutine
	// misses the cache at once.
	config, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "matching_basic.yaml")),
		util.WithOutputFormat("json"),
	)
	require.NoError(t, err)

	handler, err := newHealthHandler(config)
	require.NoError(t, err)

	const n = 16
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
			if rr.Code != http.StatusOK {
				t.Errorf("got status %d, want %d", rr.Code, http.StatusOK)
			}
		}()
	}
	wg.Wait()
}

// TestServeDeduplicatesConcurrentCacheMisses covers the cold window: when many
// probes arrive before any result is cached, they must share one validation
// run rather than each executing every check against the machine.
func TestServeDeduplicatesConcurrentCacheMisses(t *testing.T) {
	t.Parallel()

	// The run count comes from the cache-miss record, so the logger has to be
	// this test's own and has to be listening below DEBUG.
	logger, records := captureRecords(util.LevelTrace)

	config, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "matching_basic.yaml")),
		util.WithOutputFormat("json"),
		util.WithLogger(logger),
	)
	require.NoError(t, err)

	handler, err := newHealthHandler(config)
	require.NoError(t, err)

	const n = 16
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		}()
	}
	wg.Wait()

	runs := countStaleCacheRecords(records)
	require.Equal(t, 1, runs, "%d concurrent requests caused %d validation runs, want 1", n, runs)
}
