package goss

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// commandSpec writes a gossfile whose only resource runs a command, because the
// command output site is the one that proves a System carries the logger.
func commandSpec(t *testing.T, exec string, wantExitStatus int) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "goss.yaml")
	contents := "command:\n" +
		"  probe:\n" +
		"    exec: \"" + exec + "\"\n" +
		"    exit-status: " + strconv.Itoa(wantExitStatus) + "\n"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))

	return path
}

// TestLoggerReachesEveryComponent is the behavioural half of propagation: the
// logger given to NewConfig is the one every component the operation builds
// ends up emitting through.
func TestLoggerReachesEveryComponent(t *testing.T) {
	const marker = "propagation-marker"

	t.Run("ValidateConfig", func(t *testing.T) {
		logger, records := captureRecords(util.LevelTrace)

		config, err := util.NewConfig(
			util.WithSpecFile(commandSpec(t, "echo "+marker, 0)),
			util.WithOutputFormat("json"),
			util.WithResultWriter(io.Discard),
			util.WithLogger(logger),
		)
		require.NoError(t, err)

		code, err := Validate(t.Context(), config)
		require.NoError(t, err)
		require.Equal(t, 0, code)

		assertMarkerLogged(t, records(), marker)
		require.NotEmpty(t, recordsWithMessage(records(), "validation result"),
			"the OutputConfig should carry the logger too")
		require.NotEmpty(t, recordsWithMessage(records(), "validation summary"))
	})

	t.Run("ValidateResults", func(t *testing.T) {
		logger, records := captureRecords(slog.LevelDebug)

		config, err := util.NewConfig(
			util.WithSpecFile(commandSpec(t, "echo "+marker, 0)),
			util.WithLogger(logger),
		)
		require.NoError(t, err)

		results, err := ValidateResults(t.Context(), config)
		require.NoError(t, err)
		for range results { //nolint:revive // draining is the point
		}

		assertMarkerLogged(t, records(), marker)
	})

	t.Run("AddResources", func(t *testing.T) {
		logger, records := captureRecords(slog.LevelDebug)

		config, err := util.NewConfig(util.WithLogger(logger))
		require.NoError(t, err)
		// NewConfig defaults the timeout to zero, which the CLI overrides
		// per subcommand; without this the probe command times out instantly.
		config.Timeout = 10 * time.Second

		path := filepath.Join(t.TempDir(), "goss.yaml")
		require.NoError(t, AddResources(t.Context(), path, "Command", []string{"echo " + marker}, config))

		assertMarkerLogged(t, records(), marker)
	})

	t.Run("serve", func(t *testing.T) {
		logger, records := captureRecords(slog.LevelDebug)

		config, err := util.NewConfig(
			util.WithSpecFile(commandSpec(t, "echo "+marker, 0)),
			util.WithOutputFormat("json"),
			util.WithLogger(logger),
		)
		require.NoError(t, err)

		handler, err := newHealthHandler(config)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		require.Equal(t, http.StatusOK, rr.Code)

		assertMarkerLogged(t, records(), marker)
		require.NotEmpty(t, recordsWithMessage(records(), "validation summary"),
			"serve's OutputConfig should carry the logger too")
	})
}

// TestLoggerSurvivesReconstruction covers the two places a System is rebuilt
// mid-operation. Both are easy to miss, and a missed one loses logging for
// every attempt after the first.
func TestLoggerSurvivesReconstruction(t *testing.T) {
	const marker = "reconstruction-marker"

	t.Run("validate retry loop", func(t *testing.T) {
		logger, records := captureRecords(slog.LevelDebug)

		// A failing suite with a retry budget large enough for a second attempt
		// but small enough to end quickly.
		config, err := util.NewConfig(
			util.WithSpecFile(commandSpec(t, "echo "+marker, 1)),
			util.WithOutputFormat("json"),
			util.WithResultWriter(io.Discard),
			util.WithRetryTimeout(2*time.Second),
			util.WithSleep(10*time.Millisecond),
			util.WithLogger(logger),
		)
		require.NoError(t, err)

		code, _ := Validate(t.Context(), config)
		require.NotEqual(t, 0, code, "the suite is supposed to fail")

		summaries := recordsWithMessage(records(), "validation summary")
		require.Greater(t, len(summaries), 1,
			"Output runs once per attempt, so a retried run summarises more than once")
		require.GreaterOrEqual(t, len(commandOutputRecords(records(), marker)), 2,
			"the System rebuilt for the retry should carry the logger as well")
	})

	t.Run("serve cache expiry", func(t *testing.T) {
		const cacheTTL = 50 * time.Millisecond

		logger, records := captureRecords(slog.LevelDebug)

		config, err := util.NewConfig(
			util.WithSpecFile(commandSpec(t, "echo "+marker, 0)),
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

		probe()
		time.Sleep(cacheTTL + 10*time.Millisecond)
		probe()

		require.GreaterOrEqual(t, len(commandOutputRecords(records(), marker)), 2,
			"the validation run after the cache expired should log as well")
	})
}

func assertMarkerLogged(t *testing.T, records []map[string]any, marker string) {
	t.Helper()

	require.NotEmpty(t, commandOutputRecords(records, marker),
		"the subject command's output should reach the injected logger")
}

func commandOutputRecords(records []map[string]any, marker string) []map[string]any {
	var out []map[string]any
	for _, record := range recordsWithMessage(records, "command output") {
		if output, ok := record["output"].(string); ok && strings.Contains(output, marker) {
			out = append(out, record)
		}
	}
	return out
}
