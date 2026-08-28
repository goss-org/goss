package goss

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
)

// operationDrive is one exported entry point, called with a config the test
// supplies. Serve is absent deliberately: it registers on a process-global mux
// and never returns on success, so it is driven in a subprocess instead.
type operationDrive struct {
	name string
	run  func(t *testing.T, c *util.Config) error
}

// Each drive is arranged to reach at least one logging site, which is what makes
// the control half of the silence test meaningful. Which site that is differs by
// operation: the validate family logs its subject command's output, RenderJSON
// only ever warns about duplicate resources, and the add family warns when it
// declines to write an empty configuration.
func operationDrives() []operationDrive {
	return []operationDrive{
		{
			name: "Validate",
			run: func(t *testing.T, c *util.Config) error {
				c.Spec = commandFixture(t)
				_, err := Validate(c)
				return err
			},
		},
		{
			name: "ValidateResults",
			run: func(t *testing.T, c *util.Config) error {
				c.Spec = commandFixture(t)
				results, err := ValidateResults(c)
				if err != nil {
					return err
				}
				for range results { //nolint:revive // draining is the point
				}
				return nil
			},
		},
		{
			name: "ValidateConfig",
			run: func(t *testing.T, c *util.Config) error {
				c.Spec = commandFixture(t)
				gossConfig, err := getGossConfig(c)
				if err != nil {
					return err
				}
				_, err = ValidateConfig(c, gossConfig)
				return err
			},
		},
		{
			name: "RenderJSON",
			run: func(t *testing.T, c *util.Config) error {
				// RenderJSON's only log site is the duplicate-resource warning.
				c.Spec = duplicateFixture(t)
				_, err := RenderJSON(c)
				return err
			},
		},
		{
			name: "AddResources",
			run: func(t *testing.T, c *util.Config) error {
				return AddResources(filepath.Join(t.TempDir(), "goss.yaml"), "Command",
					[]string{"echo silent-marker"}, c)
			},
		},
		{
			name: "AutoAddResources",
			run: func(t *testing.T, c *util.Config) error {
				// No keys, so it declines to write and warns about it.
				return AutoAddResources(filepath.Join(t.TempDir(), "goss.yaml"), nil, c)
			},
		},
	}
}

// commandFixture writes a gossfile whose command both succeeds and produces
// output, so the validate family reaches its logging sites.
func commandFixture(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "goss.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`command:
  probe:
    exec: "echo silent-marker"
    exit-status: 0
`), 0o644))

	return path
}

// TestOperationsAreSilentWithoutALogger is what this whole change exists for:
// an embedder who supplies no logger gets no records, anywhere.
// The controls matter as much as the assertions. Every row is run twice, once
// with a logger to prove the operation does reach a logging site, and once
// without to prove the silence is real rather than an early return.
func TestOperationsAreSilentWithoutALogger(t *testing.T) {
	for _, drive := range operationDrives() {
		t.Run(drive.name+"/control", func(t *testing.T) {
			logger, records := captureRecords(util.LevelTrace)

			c := silenceConfig(t, util.WithLogger(logger))
			require.NoError(t, drive.run(t, c))

			require.NotEmpty(t, records(),
				"the control proves this operation reaches a logging site at all")
		})

		t.Run(drive.name+"/silent", func(t *testing.T) {
			global := captureGlobalLogger(t)

			c := silenceConfig(t)
			require.Nil(t, c.Logger)
			require.NoError(t, drive.run(t, c))

			require.Empty(t, global(), "no records may reach the process logger")
		})
	}
}

// TestNothingIsEmittedBelowError is driven rather than reasoned: with the level
// at ERROR nothing goss emits during a passing run appears, and the same drive
// at TRACE proves the run does produce records.
func TestNothingIsEmittedBelowError(t *testing.T) {
	for _, drive := range operationDrives() {
		t.Run(drive.name, func(t *testing.T) {
			atError, errorRecords := captureRecords(slog.LevelError)
			require.NoError(t, drive.run(t, silenceConfig(t, util.WithLogger(atError))))
			require.Empty(t, errorRecords(), "a passing run has nothing to say at ERROR")

			atTrace, traceRecords := captureRecords(util.LevelTrace)
			require.NoError(t, drive.run(t, silenceConfig(t, util.WithLogger(atTrace))))
			require.NotEmpty(t, traceRecords(), "the same drive is not silent at TRACE")
		})
	}
}

// TestGlobalLoggerIsLeftAsFound pins process state. setLogLevel used to call
// SetFlags and SetOutput on the process logger and never put them back, so a
// library consumer had their own logging reconfigured by calling goss.
// The comparison has to happen inside the window: restoring the writer first
// and comparing afterwards would erase the evidence.
func TestGlobalLoggerIsLeftAsFound(t *testing.T) {
	for _, drive := range operationDrives() {
		t.Run(drive.name, func(t *testing.T) {
			sentinel := &syncBuffer{}
			originalWriter := log.Writer()
			originalFlags := log.Flags()
			log.SetOutput(sentinel)
			log.SetFlags(log.Ldate | log.Lshortfile)
			t.Cleanup(func() {
				log.SetOutput(originalWriter)
				log.SetFlags(originalFlags)
			})

			logger, _ := captureRecords(util.LevelTrace)
			require.NoError(t, drive.run(t, silenceConfig(t, util.WithLogger(logger))))

			require.Same(t, sentinel, log.Writer(), "the writer should be untouched")
			require.Equal(t, log.Ldate|log.Lshortfile, log.Flags(), "the flags should be untouched")
			require.Empty(t, sentinel.String(), "and nothing should have been written to it")
		})
	}
}

// TestOperationsSurviveANilLogger drives the nil case. util.Config is
// documented as being usable as a literal, so a nil logger is the ordinary case
// for an embedder, and a nil *slog.Logger panics on every method that matters.
func TestOperationsSurviveANilLogger(t *testing.T) {
	for _, drive := range operationDrives() {
		t.Run(drive.name, func(t *testing.T) {
			done := make(chan struct{})
			go func() {
				defer close(done)
				// Whether the operation errors is not the point; not panicking
				// and not hanging is.
				_ = drive.run(t, silenceConfig(t))
			}()

			select {
			case <-done:
			case <-time.After(30 * time.Second):
				t.Fatal("the operation hung")
			}
		})
	}
}

func silenceConfig(t *testing.T, opts ...util.ConfigOption) *util.Config {
	t.Helper()

	c, err := util.NewConfig(append([]util.ConfigOption{
		util.WithOutputFormat("json"),
		util.WithResultWriter(io.Discard),
	}, opts...)...)
	require.NoError(t, err)

	// NewConfig leaves the command timeout at zero, which the CLI sets per
	// subcommand; without it the probe command times out before it can log.
	c.Timeout = 10 * time.Second

	return c
}
