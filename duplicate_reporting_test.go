package goss

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// duplicateFixture writes a gossfile that imports a second one redefining the
// same resource, which is the only way a duplicate reaches the merge path.
func duplicateFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	imported := filepath.Join(dir, "imported.yaml")
	require.NoError(t, os.WriteFile(imported, []byte(`command:
  duplicated-command:
    exec: "echo imported"
    exit-status: 0
`), 0o644))

	main := filepath.Join(dir, "goss.yaml")
	require.NoError(t, os.WriteFile(main, []byte(`command:
  duplicated-command:
    exec: "echo main"
    exit-status: 0
gossfile:
  imported.yaml: {}
`), 0o644))

	return main
}

// TestDuplicateResourceIsReportedThroughTheInjectedLogger covers the callback.
// The detection stays in mergeType, five frames below any config, and the
// reporting happens at the operation root that does have one.
func TestDuplicateResourceIsReportedThroughTheInjectedLogger(t *testing.T) {
	spec := duplicateFixture(t)

	t.Run("getGossConfig", func(t *testing.T) {
		logger, records := captureRecords(slog.LevelWarn)

		c, err := util.NewConfig(util.WithSpecFile(spec), util.WithLogger(logger))
		require.NoError(t, err)

		_, err = getGossConfig(c)
		require.NoError(t, err)

		assertDuplicateWarning(t, records())
	})

	t.Run("RenderJSON", func(t *testing.T) {
		logger, records := captureRecords(slog.LevelWarn)

		c, err := util.NewConfig(util.WithSpecFile(spec), util.WithLogger(logger))
		require.NoError(t, err)

		_, err = RenderJSON(c)
		require.NoError(t, err)

		assertDuplicateWarning(t, records())
	})
}

func assertDuplicateWarning(t *testing.T, records []map[string]any) {
	t.Helper()

	warnings := recordsWithMessage(records, "duplicate resource overwritten")

	require.Len(t, warnings, 1, "records: %v", records)
	require.Equal(t, "WARN", warnings[0]["level"])
	require.Equal(t, "command", warnings[0]["resource_type"],
		"the type belongs in an attribute, not interpolated into the message")
	require.Equal(t, "duplicated-command", warnings[0]["resource_id"])
}

// TestDuplicateResourceIsSilentWithoutALogger is the other half: with no logger
// supplied, nothing is written anywhere, including through the process logger
// the previous implementation used.
func TestDuplicateResourceIsSilentWithoutALogger(t *testing.T) {
	spec := duplicateFixture(t)

	c, err := util.NewConfig(util.WithSpecFile(spec))
	require.NoError(t, err)
	require.Nil(t, c.Logger)

	global := captureGlobalLogger(t)

	_, err = getGossConfig(c)
	require.NoError(t, err)

	require.Empty(t, global())
}

// TestDirectMergeIsSilent pins a deliberate behaviour change. GossConfig.Merge
// is exported and has no logger to report through, so it keeps its signature
// and stops writing to the process logger. Reporting for the operations that
// own a config is asserted above.
func TestDirectMergeIsSilent(t *testing.T) {
	first, err := ReadJSONData([]byte(`command:
  duplicated-command:
    exec: "echo first"
    exit-status: 0
`), true)
	require.NoError(t, err)

	second, err := ReadJSONData([]byte(`command:
  duplicated-command:
    exec: "echo second"
    exit-status: 0
`), true)
	require.NoError(t, err)

	global := captureGlobalLogger(t)

	first.Merge(second)

	require.Empty(t, global(), "a direct Merge has no logger and must stay silent")
	require.Equal(t, "echo second", first.Commands["duplicated-command"].Exec,
		"the merge itself must still overwrite")
}
