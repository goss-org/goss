package goss

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
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

		code, err := Validate(config)
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

		results, err := ValidateResults(config)
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
		require.NoError(t, AddResources(path, "Command", []string{"echo " + marker}, config))

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

		code, _ := Validate(config)
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

// TestEveryConstructionSiteInjectsTheLogger is the static half. The behavioural
// drives above cover the paths a test can reach; this covers the ones it
// cannot, and stops a later construction site being added without injection.
func TestEveryConstructionSiteInjectsTheLogger(t *testing.T) {
	systemNewSites := 0
	outputConfigSites := 0

	forEachShippedFile(t, func(path string, file *ast.File, fset *token.FileSet) {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				sel, ok := node.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "New" {
					return true
				}
				pkg, ok := sel.X.(*ast.Ident)
				if !ok || pkg.Name != "system" {
					return true
				}

				systemNewSites++
				require.Len(t, node.Args, 2,
					"%s: system.New should be given a logger option", fset.Position(node.Pos()))
				require.Contains(t, render(fset, node.Args[1]), "WithLogger",
					"%s: system.New should be given a logger option", fset.Position(node.Pos()))

			case *ast.CompositeLit:
				sel, ok := node.Type.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "OutputConfig" {
					return true
				}

				outputConfigSites++
				require.Contains(t, render(fset, node), "Logger:",
					"%s: an OutputConfig built by goss should carry the logger",
					fset.Position(node.Pos()))
			}
			return true
		})
	})

	// A redundant eighth construction went away with the serve conversion; if one
	// comes back, or one of the seven goes away, this is where it gets noticed.
	require.Equal(t, 7, systemNewSites)
	require.Equal(t, 2, outputConfigSites)
}

func render(fset *token.FileSet, node ast.Node) string {
	var sb strings.Builder
	_ = printer.Fprint(&sb, fset, node)

	return sb.String()
}

// forEachShippedFile parses every non-test Go file in the module.
func forEachShippedFile(t *testing.T, fn func(path string, file *ast.File, fset *token.FileSet)) {
	t.Helper()

	fset := token.NewFileSet()
	require.NoError(t, filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" || entry.Name() == "docs" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		fn(path, parsed, fset)

		return nil
	}))
}
