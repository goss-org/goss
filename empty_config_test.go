package goss

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// TestWriteJSONPreservesItsContractOnAnEmptyConfig pins the split. The exported
// function still returns nil and writes nothing, so no embedder's error
// handling changes; the unexported helper is what tells an in-package caller
// which of the two happened.
func TestWriteJSONPreservesItsContractOnAnEmptyConfig(t *testing.T) {
	// Serial: the store format is package-level state.
	setStoreFormat(YAML)

	path := filepath.Join(t.TempDir(), "goss.yaml")

	require.NoError(t, WriteJSON(path, *NewGossConfig()))
	_, err := os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist, "no file should have been written")

	written, err := writeJSON(path, *NewGossConfig())
	require.NoError(t, err)
	require.False(t, written, "the helper must report that it declined")
	_, err = os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}

// TestWriteJSONWritesANonEmptyConfig is the positive control: without it the
// test above would pass on a WriteJSON that never writes anything at all.
func TestWriteJSONWritesANonEmptyConfig(t *testing.T) {
	setStoreFormat(YAML)

	path := filepath.Join(t.TempDir(), "goss.yaml")

	gossConfig, err := ReadJSONData([]byte(`command:
  echo-hello:
    exec: "echo hello"
    exit-status: 0
`), true)
	require.NoError(t, err)

	written, err := writeJSON(path, gossConfig)
	require.NoError(t, err)
	require.True(t, written)

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(contents), "echo-hello")
}

// TestAddResourcesWarnsWhenNothingIsWritten drives the warning. The message it
// replaces was emitted from store.go through the process logger, where no level
// applied to it and no embedder could redirect it.
func TestAddResourcesWarnsWhenNothingIsWritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "goss.yaml")

	logger, records := captureRecords(slog.LevelWarn)
	c, err := util.NewConfig(util.WithLogger(logger))
	require.NoError(t, err)

	// No keys means nothing is added, so the marshalled config matches an empty
	// one and the write is declined.
	require.NoError(t, AddResources(path, "Command", nil, c))

	assertEmptyConfigWarning(t, records(), path)
	_, err = os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}

// TestAutoAddResourcesWarnsWhenNothingIsWritten is the AutoAddResources half of
// the same warning.
func TestAutoAddResourcesWarnsWhenNothingIsWritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "goss.yaml")

	logger, records := captureRecords(slog.LevelWarn)
	c, err := util.NewConfig(util.WithLogger(logger))
	require.NoError(t, err)

	require.NoError(t, AutoAddResources(path, nil, c))

	assertEmptyConfigWarning(t, records(), path)
	_, err = os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}

// TestAddRootsAreSilentWithoutALogger keeps the two new sites silent when no
// logger was supplied.
func TestAddRootsAreSilentWithoutALogger(t *testing.T) {
	c, err := util.NewConfig()
	require.NoError(t, err)

	global := captureGlobalLogger(t)

	require.NoError(t, AddResources(filepath.Join(t.TempDir(), "goss.yaml"), "Command", nil, c))
	require.NoError(t, AutoAddResources(filepath.Join(t.TempDir(), "goss.yaml"), nil, c))

	require.Empty(t, global())
}

func assertEmptyConfigWarning(t *testing.T, records []map[string]any, path string) {
	t.Helper()

	warnings := recordsWithMessage(records, "empty configuration not written")

	require.Len(t, warnings, 1, "records: %v", records)
	require.Equal(t, "WARN", warnings[0]["level"])
	require.Equal(t, path, warnings[0]["path"])
}
