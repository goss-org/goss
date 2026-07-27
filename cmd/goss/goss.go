package main

import (
	"context"
	"fmt"
	"log"
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

// converts a cli context into a goss Config
func newRuntimeConfigFromCLI(c *cli.Command) *util.Config {
	cfg := &util.Config{
		AllowInsecure:     c.Bool("insecure"),
		AnnounceToCLI:     true,
		Cache:             c.Duration("cache"),
		Debug:             c.Bool("debug"),
		LogLevel:          c.String("log-level"),
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

	return cfg
}

func timeoutFlag(value time.Duration) *cli.DurationFlag {
	return &cli.DurationFlag{
		Name:  "timeout",
		Value: value,
	}
}

func main() {
	app := &cli.Command{
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
				Action: func(ctx context.Context, c *cli.Command) error {
					fatalAlphaIfNeeded(c)
					code, err := goss.Validate(newRuntimeConfigFromCLI(c))
					if err != nil {
						color.Red(fmt.Sprintf("Error: %v\n", err))
					}
					os.Exit(code)

					return nil
				},
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
				Action: func(ctx context.Context, c *cli.Command) error {
					fatalAlphaIfNeeded(c)
					return goss.Serve(newRuntimeConfigFromCLI(c))
				},
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
				Action: func(ctx context.Context, c *cli.Command) error {
					fatalAlphaIfNeeded(c)
					j, err := goss.RenderJSON(newRuntimeConfigFromCLI(c))
					if err != nil {
						return err
					}

					fmt.Print(j)

					return nil
				},
			},
			{
				Name:    "autoadd",
				Aliases: []string{"aa"},
				Usage:   "automatically add all matching resource to the test suite",
				Action: func(ctx context.Context, c *cli.Command) error {
					fatalAlphaIfNeeded(c)
					return goss.AutoAddResources(c.String("gossfile"), c.Args().Slice(), newRuntimeConfigFromCLI(c))
				},
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
						Name:  resource.PackageResourceKey,
						Usage: "add new package",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.PackageResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.FileResourceKey,
						Usage: "add new file",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.FileResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.AddrResourceKey,
						Usage: "add new remote address:port - ex: google.com:80",
						Flags: []cli.Flag{
							timeoutFlag(500 * time.Millisecond),
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.AddResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.PortResourceKey,
						Usage: "add new listening [protocol]:port - ex: 80 or udp:123",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.PortResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.ServiceResourceKey,
						Usage: "add new service",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.ServiceResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.UserResourceKey,
						Usage: "add new user",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.UserResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.GroupResourceKey,
						Usage: "add new group",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.GroupResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.CommandResourceKey,
						Usage: "add new command",
						Flags: []cli.Flag{
							timeoutFlag(10 * time.Second),
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.CommandResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
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
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.DNSResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.ProcessResourceKey,
						Usage: "add new process name",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.ProcessResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
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
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.HTTPResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  "goss",
						Usage: "add new goss file, it will be imported from this one",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.GossFileResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))

						},
					},
					{
						Name:  resource.KernelParamResourceKey,
						Usage: "add new goss kernel param",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.KernelParamResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.MountResourceKey,
						Usage: "add new mount",
						Flags: []cli.Flag{
							timeoutFlag(1000 * time.Millisecond),
						},
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.MountResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
					{
						Name:  resource.InterfaceResourceKey,
						Usage: "add new interface",
						Action: func(ctx context.Context, c *cli.Command) error {
							fatalAlphaIfNeeded(c)
							return goss.AddResources(c.String("gossfile"), resource.InterfaceResourceName, c.Args().Slice(), newRuntimeConfigFromCLI(c))
						},
					},
				},
			},
		},
	}

	addAlphaFlagIfNeeded(app)
	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
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

func fatalAlphaIfNeeded(c *cli.Command) {
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		if c.String("use-alpha") != "1" {
			howto := map[string]string{
				"darwin":  "export GOSS_USE_ALPHA=1",
				"windows": "In cmd:        set GOSS_USE_ALPHA=1\nIn powershell: $env:GOSS_USE_ALPHA=1\nIn bash:       export GOSS_USE_ALPHA=1",
			}
			log.Printf(`Terminating.

To bypass this and use the binary anyway:

%s`, howto[runtime.GOOS])
			os.Exit(1)
		}
	}
}
