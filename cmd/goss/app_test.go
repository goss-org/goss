package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/util"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// recordedCall is one invocation of an injected goss operation.
type recordedCall struct {
	operation    string
	gossfile     string
	resourceName string
	args         []string
	config       *util.Config
}

// fakeOperations stands in for the goss package. Tests drive the production
// app, actions and runtime-config wrapper; only the operations at the far end
// are replaced, so no test needs to reach inside an Action or exit the test
// process.
type fakeOperations struct {
	mu    sync.Mutex
	calls []recordedCall
	exits []int

	validateCode int
	validateErr  error
	serveErr     error
	renderJSON   string
	renderErr    error
	autoAddErr   error
	addErr       error
}

func (f *fakeOperations) record(c recordedCall) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, c)
}

func (f *fakeOperations) recorded() []recordedCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedCall(nil), f.calls...)
}

func (f *fakeOperations) exitCodes() []int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]int(nil), f.exits...)
}

func (f *fakeOperations) operations() operations {
	return operations{
		validate: func(c *util.Config) (int, error) {
			f.record(recordedCall{operation: "validate", config: c})
			return f.validateCode, f.validateErr
		},
		serve: func(c *util.Config) error {
			f.record(recordedCall{operation: "serve", config: c})
			return f.serveErr
		},
		renderJSON: func(c *util.Config) (string, error) {
			f.record(recordedCall{operation: "render", config: c})
			return f.renderJSON, f.renderErr
		},
		autoAddResources: func(gossfile string, args []string, c *util.Config) error {
			f.record(recordedCall{operation: "autoadd", gossfile: gossfile, args: args, config: c})
			return f.autoAddErr
		},
		addResources: func(gossfile, resourceName string, args []string, c *util.Config) error {
			f.record(recordedCall{
				operation:    "add",
				gossfile:     gossfile,
				resourceName: resourceName,
				args:         args,
				config:       c,
			})
			return f.addErr
		},
		exit: func(code int) {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.exits = append(f.exits, code)
		},
	}
}

// run drives the production app with injected operations. The alpha opt-in is
// set because the platform gate runs before anything this file is testing, and
// on macOS and Windows it would otherwise end the test process.
func run(t *testing.T, fake *fakeOperations, args ...string) error {
	t.Helper()

	t.Setenv("GOSS_USE_ALPHA", "1")

	app := newApp(fake.operations())
	addAlphaFlagIfNeeded(app)
	app.Writer = io.Discard
	app.ErrWriter = io.Discard

	return app.Run(context.Background(), append([]string{"goss"}, args...))
}

// subcommandDrives are the executable actions, each in its shortest form.
func subcommandDrives() map[string][]string {
	return map[string][]string{
		"validate": {"validate"},
		"serve":    {"serve"},
		"render":   {"render"},
		"autoadd":  {"autoadd"},
		"add":      {"add", resource.PackageResourceKey, "curl"},
	}
}

// TestLogLevelAppliesToEverySubcommand drives them all. The level reaches the
// operation as a property of the injected logger's handler, which is what makes
// it work under render and autoadd, where the flag is a no-op today.
func TestLogLevelAppliesToEverySubcommand(t *testing.T) {
	levels := map[string]slog.Level{
		"TRACE": util.LevelTrace,
		"DEBUG": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"WARN":  slog.LevelWarn,
		"ERROR": slog.LevelError,
	}

	for name, drive := range subcommandDrives() {
		for spelling, want := range levels {
			for _, cased := range []string{spelling, strings.ToLower(spelling)} {
				t.Run(name+"/"+cased, func(t *testing.T) {
					fake := &fakeOperations{}

					args := append([]string{"--log-level", cased}, drive...)
					require.NoError(t, run(t, fake, args...))

					calls := fake.recorded()
					require.Len(t, calls, 1)
					assertLevel(t, calls[0].config, want)
				})
			}
		}
	}
}

// TestLogLevelFromEnvironment covers the GOSS_LOGLEVEL half of that. An
// environment value is indistinguishable from an explicit flag by design, so
// the same resolution has to happen.
func TestLogLevelFromEnvironment(t *testing.T) {
	for name, drive := range subcommandDrives() {
		t.Run(name, func(t *testing.T) {
			fake := &fakeOperations{}

			t.Setenv("GOSS_LOGLEVEL", "trace")
			require.NoError(t, run(t, fake, drive...))

			calls := fake.recorded()
			require.Len(t, calls, 1)
			assertLevel(t, calls[0].config, util.LevelTrace)
		})
	}
}

// TestDefaultLogLevel pins the CLI's own default, which stays INFO.
func TestDefaultLogLevel(t *testing.T) {
	fake := &fakeOperations{}

	require.NoError(t, run(t, fake, "serve"))

	calls := fake.recorded()
	require.Len(t, calls, 1)
	assertLevel(t, calls[0].config, slog.LevelInfo)
}

func assertLevel(t *testing.T, c *util.Config, want slog.Level) {
	t.Helper()

	require.NotNil(t, c)
	require.NotNil(t, c.Logger, "every action should be handed a logger")
	require.True(t, c.Logger.Enabled(context.Background(), want),
		"records at %s should be emitted", want)
	require.False(t, c.Logger.Enabled(context.Background(), want-1),
		"records below %s should be suppressed", want)
}

// TestInvalidLogLevelStopsEveryOperation pins the rejection: an unsupported
// value fails during runtime config construction, before the subcommand's
// operation runs. Today render and autoadd would proceed, because they never
// validated it.
func TestInvalidLogLevelStopsEveryOperation(t *testing.T) {
	for name, drive := range subcommandDrives() {
		t.Run(name+"/argv", func(t *testing.T) {
			fake := &fakeOperations{}

			args := append([]string{"--log-level", "VERBOSE"}, drive...)
			err := run(t, fake, args...)

			require.Error(t, err)
			require.Contains(t, err.Error(), "VERBOSE")
			require.Empty(t, fake.recorded(), "the operation should not have run")
			require.Empty(t, fake.exitCodes(), "the process should not have been exited")
		})

		t.Run(name+"/environment", func(t *testing.T) {
			fake := &fakeOperations{}

			t.Setenv("GOSS_LOGLEVEL", "VERBOSE")
			err := run(t, fake, drive...)

			require.Error(t, err)
			require.Contains(t, err.Error(), "VERBOSE")
			require.Empty(t, fake.recorded(), "the operation should not have run")
		})
	}
}

// TestValidateExitAdapter pins the one action that does not simply return: it
// reports the error and exits with the code goss.Validate returned.
func TestValidateExitAdapter(t *testing.T) {
	t.Run("failing suite", func(t *testing.T) {
		fake := &fakeOperations{validateCode: 1}

		require.NoError(t, run(t, fake, "validate"))
		require.Equal(t, []int{1}, fake.exitCodes())
	})

	t.Run("error is reported and its code still exits", func(t *testing.T) {
		fake := &fakeOperations{validateCode: 78, validateErr: errors.New("boom")}

		require.NoError(t, run(t, fake, "validate"))
		require.Equal(t, []int{78}, fake.exitCodes())
	})
}

// TestRenderRunsItsOperation keeps the render action wired to the operation it
// names. Where its output goes is asserted against the real binary in
// TestStdoutCarriesNoLogRecords, because a package that swaps os.Stdout in
// process cannot also run its tests in parallel.
func TestRenderRunsItsOperation(t *testing.T) {
	fake := &fakeOperations{renderJSON: "rendered-marker"}

	require.NoError(t, run(t, fake, "render"))

	calls := fake.recorded()
	require.Len(t, calls, 1)
	require.Equal(t, "render", calls[0].operation)
}

// TestRenderPropagatesItsError pins that a failing render still surfaces.
func TestRenderPropagatesItsError(t *testing.T) {
	fake := &fakeOperations{renderErr: errors.New("render-boom")}

	err := run(t, fake, "render")
	require.ErrorContains(t, err, "render-boom")
}

// TestAddActionsPassTheirResourceName covers all 16 nested add actions. The
// shared dependency must not cost them their own resource name.
func TestAddActionsPassTheirResourceName(t *testing.T) {
	want := map[string]string{
		resource.PackageResourceKey:     resource.PackageResourceName,
		resource.FileResourceKey:        resource.FileResourceName,
		resource.AddrResourceKey:        resource.AddResourceName,
		resource.PortResourceKey:        resource.PortResourceName,
		resource.ServiceResourceKey:     resource.ServiceResourceName,
		resource.UserResourceKey:        resource.UserResourceName,
		resource.GroupResourceKey:       resource.GroupResourceName,
		resource.CommandResourceKey:     resource.CommandResourceName,
		resource.DNSResourceKey:         resource.DNSResourceName,
		resource.ProcessResourceKey:     resource.ProcessResourceName,
		resource.HTTPResourceKey:        resource.HTTPResourceName,
		"goss":                          resource.GossFileResourceName,
		resource.KernelParamResourceKey: resource.KernelParamResourceName,
		resource.MountResourceKey:       resource.MountResourceName,
		resource.InterfaceResourceKey:   resource.InterfaceResourceName,
		resource.RegistryResourceKey:    resource.RegistryResourceName,
	}

	// The count is asserted against the app itself so that a new add
	// subcommand cannot be introduced without a row here.
	add := findCommand(t, newApp((&fakeOperations{}).operations()), "add")
	require.Len(t, add.Commands, len(want))

	for key, resourceName := range want {
		t.Run(key, func(t *testing.T) {
			fake := &fakeOperations{}

			require.NoError(t, run(t, fake, "add", key, "subject"))

			calls := fake.recorded()
			require.Len(t, calls, 1)
			require.Equal(t, "add", calls[0].operation)
			require.Equal(t, resourceName, calls[0].resourceName)
			require.Equal(t, []string{"subject"}, calls[0].args)
			require.Equal(t, "./goss.yaml", calls[0].gossfile)
			assertLevel(t, calls[0].config, slog.LevelInfo)
		})
	}
}

// TestAutoAddPassesItsArguments keeps autoadd's own shape, which differs from
// add's by having no resource name.
func TestAutoAddPassesItsArguments(t *testing.T) {
	fake := &fakeOperations{}

	require.NoError(t, run(t, fake, "--gossfile", "custom.yaml", "autoadd", "subject"))

	calls := fake.recorded()
	require.Len(t, calls, 1)
	require.Equal(t, "autoadd", calls[0].operation)
	require.Equal(t, "custom.yaml", calls[0].gossfile)
	require.Equal(t, []string{"subject"}, calls[0].args)
}

func findCommand(t *testing.T, app *cli.Command, name string) *cli.Command {
	t.Helper()

	for _, c := range app.Commands {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("command %q not found", name)
	return nil
}
