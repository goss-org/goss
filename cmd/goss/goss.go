package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/goss-org/goss"
	"github.com/goss-org/goss/outputs"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

// operations is the set of goss entry points the CLI drives, plus the process
// exit it needs. Holding them in one value lets the app be built and driven with
// fakes in a test, without replacing an Action or ending the test process, and
// without any production branch that exists only to enable that.
type operations struct {
	validate         func(*util.Config) (int, error)
	serve            func(*util.Config) error
	renderJSON       func(*util.Config) (string, error)
	autoAddResources func(string, []string, *util.Config) error
	addResources     func(string, string, []string, *util.Config) error
	exit             func(int)
}

// defaultOperations is what the goss binary runs with.
func defaultOperations() operations {
	return operations{
		validate:         goss.Validate,
		serve:            goss.Serve,
		renderJSON:       goss.RenderJSON,
		autoAddResources: goss.AutoAddResources,
		addResources:     goss.AddResources,
		exit:             os.Exit,
	}
}

// action builds the Action for an executable subcommand. Every one of them does
// the same two things first, in this order: the platform alpha gate, which may
// end the process, and then runtime configuration construction, whose error is
// returned without the operation being invoked. Alpha rejection therefore wins
// when both the platform and the log level are unacceptable.
func (o operations) action(fn func(context.Context, *cli.Command, *util.Config) error) func(context.Context, *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		fatalAlphaIfNeeded(c)

		cfg, err := newRuntimeConfigFromCLI(c)
		if err != nil {
			return err
		}

		return fn(ctx, c, cfg)
	}
}

// addAction builds the Action for one of the add subcommands. They differ only
// in the resource name they name, which stays with the subcommand that owns it.
func (o operations) addAction(resourceName string) func(context.Context, *cli.Command) error {
	return o.action(func(ctx context.Context, c *cli.Command, cfg *util.Config) error {
		return o.addResources(c.String("gossfile"), resourceName, c.Args().Slice(), cfg)
	})
}

// newRuntimeConfigFromCLI converts a cli context into a goss Config, including
// the logger goss writes through. The level belongs to the handler, so it is
// resolved here, at the edge, rather than being carried into the library as a
// string for the library to interpret.
func newRuntimeConfigFromCLI(c *cli.Command) (*util.Config, error) {
	level, err := parseLogLevel(c.String("log-level"))
	if err != nil {
		return nil, err
	}

	cfg := &util.Config{
		AllowInsecure: c.Bool("insecure"),
		AnnounceToCLI: true,
		Cache:         c.Duration("cache"),
		Debug:         c.Bool("debug"),
		Logger:        slog.New(newCLIHandler(os.Stderr, level)),

		Endpoint:          c.String("endpoint"),
		FormatOptions:     c.StringSlice("format-options"),
		IgnoreList:        c.StringSlice("exclude-attr"),
		ListenAddress:     c.String("listen-addr"),
		MaxConcurrent:     c.Int("max-concurrent"),
		NoFollowRedirects: c.Bool("no-follow-redirects"),
		OutputFormat:      c.String("format"),
		PackageManager:    c.String("package"),
		Password:          c.String("password"),
		Proxy:             c.String("proxy"),
		RetryTimeout:      c.Duration("retry-timeout"),
		Server:            c.String("server"),
		Sleep:             c.Duration("sleep"),
		Spec:              c.String("gossfile"),
		Timeout:           c.Duration("timeout"),
		Username:          c.String("username"),
		VarsInline:        c.String("vars-inline"),
		VarsFiles:         c.StringSlice("vars"),
	}

	if c.Bool("no-color") {
		util.WithNoColor()(cfg)
	}

	if c.Bool("color") {
		util.WithColor()(cfg)
	}

	return cfg, nil
}

func timeoutFlag(value time.Duration) *cli.DurationFlag {
	return &cli.DurationFlag{
		Name:  "timeout",
		Value: value,
	}
}

func main() {
	app := newApp(defaultOperations())

	addAlphaFlagIfNeeded(app)
	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// newApp builds the goss CLI. ops is the far end of every action, so a test can
// drive the real app, flags and actions against fakes.
func newApp(ops operations) *cli.Command {
	return &cli.Command{
		EnableShellCompletion: true,
		Version:               util.Version,
		Name:                  "goss",
		Usage:                 "Quick and Easy server validation",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"loglevel", "L", "l"},
				Value:   "INFO",
				Usage:   "Goss log verbosity level",
				Sources: cli.EnvVars("GOSS_LOGLEVEL"),
			},
			&cli.StringFlag{
				Name:    "gossfile",
				Aliases: []string{"g"},
				Value:   "./goss.yaml",
				Usage:   "Goss file to read from / write to",
				Sources: cli.EnvVars("GOSS_FILE"),
			},
			&cli.StringSliceFlag{
				Name:    "vars",
				Usage:   "json/yaml file containing variables for template",
				Sources: cli.EnvVars("GOSS_VARS"),
			},
			&cli.StringFlag{
				Name:    "vars-inline",
				Usage:   "json/yaml string containing variables for template (overwrites vars)",
				Sources: cli.EnvVars("GOSS_VARS_INLINE"),
			},
			&cli.StringFlag{
				Name:  "package",
				Usage: fmt.Sprintf("Package type to use [%s]", strings.Join(system.SupportedPackageManagers(), ", ")),
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "validate",
				Aliases: []string{"v"},
				Usage:   "Validate system",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Value:   "rspecish",
						Usage:   fmt.Sprintf("Format to output in, valid options: %s", outputs.Outputers()),
						Sources: cli.EnvVars("GOSS_FMT"),
					},
					&cli.StringSliceFlag{
						Name:    "format-options",
						Aliases: []string{"o"},
						Usage:   fmt.Sprintf("Extra options passed to the formatter, valid options: %s", outputs.FormatOptions()),
						Sources: cli.EnvVars("GOSS_FMT_OPTIONS"),
					},
					&cli.BoolFlag{
						Name:    "color",
						Usage:   "Force color on",
						Sources: cli.EnvVars("GOSS_COLOR"),
					},
					&cli.BoolFlag{
						Name:    "no-color",
						Usage:   "Force color off",
						Sources: cli.EnvVars("GOSS_NOCOLOR"),
					},
					&cli.DurationFlag{
						Name:    "sleep",
						Aliases: []string{"s"},
						Usage:   "Time to sleep between retries, only active when -r is set",
						Value:   1 * time.Second,
						Sources: cli.EnvVars("GOSS_SLEEP"),
					},
					&cli.DurationFlag{
						Name:    "retry-timeout",
						Aliases: []string{"r"},
						Usage:   "Retry on failure so long as elapsed + sleep time is less than this",
						Value:   0,
						Sources: cli.EnvVars("GOSS_RETRY_TIMEOUT"),
					},
					&cli.IntFlag{
						Name:    "max-concurrent",
						Usage:   "Max number of tests to run concurrently",
						Value:   50,
						Sources: cli.EnvVars("GOSS_MAX_CONCURRENT"),
					},
				},
				Action: ops.action(func(ctx context.Context, c *cli.Command, cfg *util.Config) error {
					code, err := ops.validate(cfg)
					if err != nil {
						color.Red(fmt.Sprintf("Error: %v\n", err))
					}
					ops.exit(code)

					return nil
				}),
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Serve a health endpoint",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Value:   "rspecish",
						Usage:   fmt.Sprintf("Format to output in, valid options: %s", outputs.Outputers()),
						Sources: cli.EnvVars("GOSS_FMT"),
					},
					&cli.StringSliceFlag{
						Name:    "format-options",
						Aliases: []string{"o"},
						Usage:   fmt.Sprintf("Extra options passed to the formatter, valid options: %s", outputs.FormatOptions()),
						Sources: cli.EnvVars("GOSS_FMT_OPTIONS"),
					},
					&cli.DurationFlag{
						Name:    "cache",
						Aliases: []string{"c"},
						Usage:   "Time to cache the results",
						Value:   5 * time.Second,
						Sources: cli.EnvVars("GOSS_CACHE"),
					},
					&cli.StringFlag{
						Name:    "listen-addr",
						Aliases: []string{"l"},
						Value:   ":8080",
						Usage:   "Address to listen on [ip]:port",
						Sources: cli.EnvVars("GOSS_LISTEN"),
					},
					&cli.StringFlag{
						Name:    "endpoint",
						Aliases: []string{"e"},
						Value:   "/healthz",
						Usage:   "Endpoint to expose",
						Sources: cli.EnvVars("GOSS_ENDPOINT"),
					},
					&cli.IntFlag{
						Name:    "max-concurrent",
						Usage:   "Max number of tests to run concurrently",
						Value:   50,
						Sources: cli.EnvVars("GOSS_MAX_CONCURRENT"),
					},
				},
				Action: ops.action(func(ctx context.Context, c *cli.Command, cfg *util.Config) error {
					return ops.serve(cfg)
				}),
			},
			{
				Name:    "render",
				Aliases: []string{"r"},
				Usage:   "render gossfile after imports",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						Usage:   "Print debugging info when rendering",
					},
				},
				Action: ops.action(func(ctx context.Context, c *cli.Command, cfg *util.Config) error {
					j, err := ops.renderJSON(cfg)
					if err != nil {
						return err
					}

					fmt.Print(j)

					return nil
				}),
			},
			{
				Name:    "autoadd",
				Aliases: []string{"aa"},
				Usage:   "automatically add all matching resource to the test suite",
				Action: ops.action(func(ctx context.Context, c *cli.Command, cfg *util.Config) error {
					return ops.autoAddResources(c.String("gossfile"), c.Args().Slice(), cfg)
				}),
			},
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a resource to the test suite",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:  "exclude-attr",
						Usage: "Exclude the following attributes when adding a new resource",
					},
				},
				Commands: []*cli.Command{
					{
						Name:   resource.PackageResourceKey,
						Usage:  "add new package",
						Action: ops.addAction(resource.PackageResourceName),
					},
					{
						Name:   resource.FileResourceKey,
						Usage:  "add new file",
						Action: ops.addAction(resource.FileResourceName),
					},
					{
						Name:  resource.AddrResourceKey,
						Usage: "add new remote address:port - ex: google.com:80",
						Flags: []cli.Flag{
							timeoutFlag(500 * time.Millisecond),
						},
						Action: ops.addAction(resource.AddResourceName),
					},
					{
						Name:   resource.PortResourceKey,
						Usage:  "add new listening [protocol]:port - ex: 80 or udp:123",
						Action: ops.addAction(resource.PortResourceName),
					},
					{
						Name:   resource.ServiceResourceKey,
						Usage:  "add new service",
						Action: ops.addAction(resource.ServiceResourceName),
					},
					{
						Name:   resource.UserResourceKey,
						Usage:  "add new user",
						Action: ops.addAction(resource.UserResourceName),
					},
					{
						Name:   resource.GroupResourceKey,
						Usage:  "add new group",
						Action: ops.addAction(resource.GroupResourceName),
					},
					{
						Name:  resource.CommandResourceKey,
						Usage: "add new command",
						Flags: []cli.Flag{
							timeoutFlag(10 * time.Second),
						},
						Action: ops.addAction(resource.CommandResourceName),
					},
					{
						Name:  resource.DNSResourceKey,
						Usage: "add new dns",
						Flags: []cli.Flag{
							timeoutFlag(500 * time.Millisecond),
							&cli.StringFlag{
								Name:  "server",
								Usage: "The IP address of a DNS server to query",
							},
						},
						Action: ops.addAction(resource.DNSResourceName),
					},
					{
						Name:   resource.ProcessResourceKey,
						Usage:  "add new process name",
						Action: ops.addAction(resource.ProcessResourceName),
					},
					{
						Name:  resource.HTTPResourceKey,
						Usage: "add new http",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:    "insecure",
								Aliases: []string{"k"},
							},
							&cli.BoolFlag{
								Name:    "no-follow-redirects",
								Aliases: []string{"r"},
							},
							timeoutFlag(5 * time.Second),
							&cli.StringFlag{
								Name:    "username",
								Aliases: []string{"u"},
								Usage:   "Username for basic auth",
							},
							&cli.StringFlag{
								Name:    "password",
								Aliases: []string{"p"},
								Usage:   "Password for basic auth",
							},
							&cli.StringFlag{
								Name:    "proxy",
								Aliases: []string{"x"},
								Usage:   "Proxy server to use. e.g. http://10.0.0.2:8080",
							},
						},
						Action: ops.addAction(resource.HTTPResourceName),
					},
					{
						Name:   "goss",
						Usage:  "add new goss file, it will be imported from this one",
						Action: ops.addAction(resource.GossFileResourceName),
					},
					{
						Name:   resource.KernelParamResourceKey,
						Usage:  "add new goss kernel param",
						Action: ops.addAction(resource.KernelParamResourceName),
					},
					{
						Name:  resource.MountResourceKey,
						Usage: "add new mount",
						Flags: []cli.Flag{
							timeoutFlag(1000 * time.Millisecond),
						},
						Action: ops.addAction(resource.MountResourceName),
					},
					{
						Name:   resource.InterfaceResourceKey,
						Usage:  "add new interface",
						Action: ops.addAction(resource.InterfaceResourceName),
					},
					{
						Name:   resource.RegistryResourceKey,
						Usage:  "add new registry key",
						Action: ops.addAction(resource.RegistryResourceName),
					},
				},
			},
		},
	}
}

func addAlphaFlagIfNeeded(cmd *cli.Command) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		cmd.Flags = append(cmd.Flags, &cli.StringFlag{
			Name:    "use-alpha",
			Usage:   "goss on macOS/Windows is alpha-quality. Set to 1 to use anyway.",
			Sources: cli.EnvVars("GOSS_USE_ALPHA"),
			Value:   "0",
		})
	}
}

// alphaRejected is the platform gate's decision, as a pure function of the OS
// name and the opt-in value. Keeping the decision separate from the process exit
// is what makes the darwin and windows behaviour testable from any host.
func alphaRejected(goos, useAlpha string) bool {
	if goos != "darwin" && goos != "windows" {
		return false
	}
	return useAlpha != "1"
}

// alphaMessage is the complete diagnostic the gate prints, including the
// platform-specific bypass instructions.
func alphaMessage(goos string) string {
	howto := map[string]string{
		"darwin":  "export GOSS_USE_ALPHA=1",
		"windows": "In cmd:        set GOSS_USE_ALPHA=1\nIn powershell: $env:GOSS_USE_ALPHA=1\nIn bash:       export GOSS_USE_ALPHA=1",
	}

	return fmt.Sprintf(`Terminating.

To bypass this and use the binary anyway:

%s`, howto[goos])
}

// fatalAlphaIfNeeded runs before anything else an action does, including runtime
// configuration construction, so no logger exists yet. It stays on the standard
// log package for that reason.
func fatalAlphaIfNeeded(c *cli.Command) {
	if alphaRejected(runtime.GOOS, c.String("use-alpha")) {
		log.Print(alphaMessage(runtime.GOOS))
		os.Exit(1)
	}
}
