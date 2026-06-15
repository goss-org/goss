package goss

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jsonOutput is the minimal subset of the JSON response we care about
// when asserting on which tests ran.
type jsonOutput struct {
	Results []struct {
		Resource     string `json:"resource-id"`
		ResourceType string `json:"resource-type"`
		Property     string `json:"property"`
		Skipped      bool   `json:"skipped"`
	} `json:"results"`
	Summary struct {
		TestCount    int `json:"test-count"`
		SkippedCount int `json:"skipped-count"`
		FailedCount  int `json:"failed-count"`
	} `json:"summary"`
}

func makeMarkRequest(t *testing.T, query string) *http.Request {
	t.Helper()
	url := "/healthz"
	if query != "" {
		url += "?" + query
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	return req
}

func decodeJSON(t *testing.T, body *bytes.Buffer) jsonOutput {
	t.Helper()
	var out jsonOutput
	require.NoError(t, json.Unmarshal(body.Bytes(), &out))
	return out
}

// TestServeWithMarksQueryParam verifies that ?marks=<list> filters tests to
// only those bearing one of the supplied marks.
func TestServeWithMarksQueryParam(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithOutputFormat("json"),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	hh, err := newHealthHandler(cfg)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	hh.ServeHTTP(rr, makeMarkRequest(t, "marks=critical"))

	require.Equal(t, http.StatusOK, rr.Code)
	out := decodeJSON(t, rr.Body)

	// We expect 6 result rows (3 commands * 2 properties), 4 skipped (slow + unmarked).
	assert.Equal(t, 6, out.Summary.TestCount)
	assert.Equal(t, 4, out.Summary.SkippedCount)
}

// TestServeWithExcludeMarksQueryParam verifies that ?exclude-marks=<list>
// skips tests that bear any of the supplied marks.
func TestServeWithExcludeMarksQueryParam(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithOutputFormat("json"),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	hh, err := newHealthHandler(cfg)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	hh.ServeHTTP(rr, makeMarkRequest(t, "exclude-marks=slow"))

	require.Equal(t, http.StatusOK, rr.Code)
	out := decodeJSON(t, rr.Body)

	// Only the slow command should be skipped (2 properties), critical + unmarked run.
	assert.Equal(t, 6, out.Summary.TestCount)
	assert.Equal(t, 2, out.Summary.SkippedCount)
}

// TestServeWithCombinedMarksQueryParams verifies inclusion is applied first,
// then exclusion.
func TestServeWithCombinedMarksQueryParams(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithOutputFormat("json"),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	hh, err := newHealthHandler(cfg)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	hh.ServeHTTP(rr, makeMarkRequest(t, "marks=critical,slow&exclude-marks=slow"))

	require.Equal(t, http.StatusOK, rr.Code)
	out := decodeJSON(t, rr.Body)

	// Include both critical and slow, then exclude slow. Only critical should run (2 properties).
	// Skipped: slow + unmarked = 4 properties.
	assert.Equal(t, 6, out.Summary.TestCount)
	assert.Equal(t, 4, out.Summary.SkippedCount)
}

// TestServeCacheIsolationPerMarks verifies that different mark combinations
// use distinct cache keys, preventing cross-contamination of cached results.
func TestServeCacheIsolationPerMarks(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithOutputFormat("json"),
		util.WithCache(5*time.Second),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	hh, err := newHealthHandler(cfg)
	require.NoError(t, err)

	// Subtests are NOT parallel within this test: they deliberately share
	// cache state and must run sequentially.
	t.Run("first request: no marks", func(t *testing.T) {
		rr := httptest.NewRecorder()
		hh.ServeHTTP(rr, makeMarkRequest(t, ""))
		require.Equal(t, http.StatusOK, rr.Code)
		out := decodeJSON(t, rr.Body)
		assert.Equal(t, 0, out.Summary.SkippedCount, "no marks: nothing should be skipped")
		assert.Contains(t, tl.String(), "Stale cache[res]", "first request should miss cache")
		tl.Reset()
	})

	t.Run("second request: ?marks=critical does not return cached unfiltered result", func(t *testing.T) {
		rr := httptest.NewRecorder()
		hh.ServeHTTP(rr, makeMarkRequest(t, "marks=critical"))
		require.Equal(t, http.StatusOK, rr.Code)
		out := decodeJSON(t, rr.Body)
		// Critical filter: slow + unmarked skipped (4 properties)
		assert.Equal(t, 4, out.Summary.SkippedCount, "must apply mark filter, not return cached unfiltered result")
		assert.Contains(t, tl.String(), "res:include=critical", "should use mark-specific cache key")
		tl.Reset()
	})

	t.Run("third request: same marks hits cache", func(t *testing.T) {
		rr := httptest.NewRecorder()
		hh.ServeHTTP(rr, makeMarkRequest(t, "marks=critical"))
		require.Equal(t, http.StatusOK, rr.Code)
		assert.NotContains(t, tl.String(), "Stale cache", "second identical request should hit cache")
		assert.Contains(t, tl.String(), "Returning cached[res:include=critical]")
		tl.Reset()
	})

	t.Run("fourth request: different marks miss cache", func(t *testing.T) {
		rr := httptest.NewRecorder()
		hh.ServeHTTP(rr, makeMarkRequest(t, "exclude-marks=slow"))
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, tl.String(), "Stale cache[res:exclude=slow]", "different mark combo should miss cache")
		tl.Reset()
	})
}

// TestServeQueryParamsOverrideConfig verifies that query parameters take
// precedence over server-level config marks.
func TestServeQueryParamsOverrideConfig(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithOutputFormat("json"),
		util.WithIncludeMarks("slow"), // server-level: only slow
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	hh, err := newHealthHandler(cfg)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	// Override with query param: only critical
	hh.ServeHTTP(rr, makeMarkRequest(t, "marks=critical"))

	require.Equal(t, http.StatusOK, rr.Code)
	out := decodeJSON(t, rr.Body)
	// Critical filter: slow + unmarked skipped (4 properties)
	assert.Equal(t, 4, out.Summary.SkippedCount,
		"query parameter marks must override config marks")

	// And the server-level config should not have been mutated
	assert.Equal(t, []string{"slow"}, cfg.IncludeMarks,
		"per-request mark filter must not mutate shared config")
}

// TestServeQueryParamsConcurrent verifies that concurrent requests with
// different mark combinations do not race on shared state. This is the
// regression test for the Skip-state mutation issue documented in the
// marks-feature design doc (section "Resource Skip State Mutation Across
// Requests"); without the snapshot/restore of skip flags under gossMu,
// request N's skip decisions would leak into request N+1.
func TestServeQueryParamsConcurrent(t *testing.T) {
	t.Parallel()
	tl := util.NewTestLogger(t)

	cfg, err := util.NewConfig(
		util.WithSpecFile(filepath.Join("testdata", "marks.goss.yaml")),
		util.WithOutputFormat("json"),
		util.WithLogger(tl),
	)
	require.NoError(t, err)

	hh, err := newHealthHandler(cfg)
	require.NoError(t, err)

	const N = 20
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		query := "marks=critical"
		expected := 4
		if i%2 == 0 {
			query = "exclude-marks=slow"
			expected = 2
		}
		go func(query string, expected int) {
			defer wg.Done()
			rr := httptest.NewRecorder()
			hh.ServeHTTP(rr, makeMarkRequest(t, query))
			if rr.Code != http.StatusOK {
				t.Errorf("query=%q got status %d", query, rr.Code)
				return
			}
			out := decodeJSON(t, rr.Body)
			if out.Summary.SkippedCount != expected {
				t.Errorf("query=%q got SkippedCount=%d want %d", query, out.Summary.SkippedCount, expected)
			}
		}(query, expected)
	}
	wg.Wait()

	// Both cache keys should appear in logs.
	logs := tl.String()
	assert.True(t, strings.Contains(logs, "res:include=critical") || strings.Contains(logs, "Returning cached[res:include=critical]"))
	assert.True(t, strings.Contains(logs, "res:exclude=slow") || strings.Contains(logs, "Returning cached[res:exclude=slow]"))
}

// TestServeMarkFilterCacheKey verifies the cache key generation for various
// mark filter combinations.
func TestServeMarkFilterCacheKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		mf   markFilter
		want string
	}{
		{"empty", markFilter{}, "res"},
		{"include only", markFilter{includeMarks: []string{"critical"}}, "res:include=critical"},
		{"exclude only", markFilter{excludeMarks: []string{"slow"}}, "res:exclude=slow"},
		{"both", markFilter{includeMarks: []string{"a", "b"}, excludeMarks: []string{"c"}}, "res:include=a,b:exclude=c"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.mf.cacheKey(); got != tc.want {
				t.Errorf("cacheKey() = %q, want %q", got, tc.want)
			}
		})
	}
}
