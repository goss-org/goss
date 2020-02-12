package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aelsabbahy/goss"
	"github.com/aelsabbahy/goss/outputs"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

var version string

// converts a cli context into a goss RuntimeConfig
func newRuntimeConfigFromCLI(c *cli.Context) *goss.RuntimeConfig {
	bp := func(b bool) *bool { return &b }

	cfg := &goss.RuntimeConfig{
		FormatOptions:     c.StringSlice("format-options"),
		Vars:              c.GlobalString("vars"),
		VarsInline:        c.GlobalString("vars-inline"),
		Spec:              c.GlobalString("gossfile"),
		Sleep:             c.Duration("sleep"),
		RetryTimeout:      c.Duration("retry-timeout"),
		Timeout:           c.Duration("timeout"),
		Cache:             c.Duration("cache"),
		MaxConcurrent:     c.Int("max-concurrent"),
		OutputFormat:      c.String("format"),
		PackageManager:    c.GlobalString("package"),
		Endpoint:          c.String("endpoint"),
		ListenAddress:     c.String("listen-addr"),
		ExcludeAttributes: c.GlobalStringSlice("exclude-attr"),
		Insecure:          c.Bool("insecure"),
		NoFollowRedirects: c.Bool("no-follow-redirects"),
		Server:            c.String("server"),
		Username:          c.String("username"),
		Password:          c.String("password"),
		Debug:             c.Bool("debug"),
	}

	if c.Bool("no-color") {
		cfg.NoColor = bp(true)
	}

	if c.Bool("color") {
		cfg.NoColor = bp(true)
	}

	return cfg
}

func main() {
	startTime := time.Now()
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = version
	app.Name = "goss"
	app.Usage = "Quick and Easy server validation"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "gossfile, g",
			Value:  "./goss.yaml",
			Usage:  "Goss file to read from / write to",
			EnvVar: "GOSS_FILE",
		},
		cli.StringFlag{
			Name:   "vars",
			Usage:  "json/yaml file containing variables for template",
			EnvVar: "GOSS_VARS",
		},
		cli.StringFlag{
			Name:   "vars-inline",
			Usage:  "json/yaml string containing variables for template (overwrites vars)",
			EnvVar: "GOSS_VARS_INLINE",
		},
		cli.StringFlag{
			Name:  "package",
			Usage: "Package type to use [rpm, deb, apk, pacman]",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "validate",
			Aliases: []string{"v"},
			Usage:   "Validate system",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "format, f",
					Value:  "rspecish",
					Usage:  fmt.Sprintf("Format to output in, valid options: %s", outputs.Outputers()),
					EnvVar: "GOSS_FMT",
				},
				cli.StringSliceFlag{
					Name:   "format-options, o",
					Usage:  fmt.Sprintf("Extra options passed to the formatter, valid options: %s", outputs.FormatOptions()),
					EnvVar: "GOSS_FMT_OPTIONS",
				},
				cli.BoolFlag{
					Name:   "color",
					Usage:  "Force color on",
					EnvVar: "GOSS_COLOR",
				},
				cli.BoolFlag{
					Name:   "no-color",
					Usage:  "Force color off",
					EnvVar: "GOSS_NOCOLOR",
				},
				cli.DurationFlag{
					Name:   "sleep,s",
					Usage:  "Time to sleep between retries, only active when -r is set",
					Value:  1 * time.Second,
					EnvVar: "GOSS_SLEEP",
				},
				cli.DurationFlag{
					Name:   "retry-timeout,r",
					Usage:  "Retry on failure so long as elapsed + sleep time is less than this",
					Value:  0,
					EnvVar: "GOSS_RETRY_TIMEOUT",
				},
				cli.IntFlag{
					Name:   "max-concurrent",
					Usage:  "Max number of tests to run concurrently",
					Value:  50,
					EnvVar: "GOSS_MAX_CONCURRENT",
				},
			},
			Action: func(c *cli.Context) error {
				code, err := goss.Validate(newRuntimeConfigFromCLI(c), startTime)
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
				cli.StringFlag{
					Name:   "format, f",
					Value:  "rspecish",
					Usage:  fmt.Sprintf("Format to output in, valid options: %s", outputs.Outputers()),
					EnvVar: "GOSS_FMT",
				},
				cli.StringSliceFlag{
					Name:   "format-options, o",
					Usage:  fmt.Sprintf("Extra options passed to the formatter, valid options: %s", outputs.FormatOptions()),
					EnvVar: "GOSS_FMT_OPTIONS",
				},
				cli.DurationFlag{
					Name:   "cache,c",
					Usage:  "Time to cache the results",
					Value:  5 * time.Second,
					EnvVar: "GOSS_CACHE",
				},
				cli.StringFlag{
					Name:   "listen-addr,l",
					Value:  ":8080",
					Usage:  "Address to listen on [ip]:port",
					EnvVar: "GOSS_LISTEN",
				},
				cli.StringFlag{
					Name:   "endpoint,e",
					Value:  "/healthz",
					Usage:  "Endpoint to expose",
					EnvVar: "GOSS_ENDPOINT",
				},
				cli.IntFlag{
					Name:   "max-concurrent",
					Usage:  "Max number of tests to run concurrently",
					Value:  50,
					EnvVar: "GOSS_MAX_CONCURRENT",
				},
			},
			Action: func(c *cli.Context) error {
				goss.Serve(newRuntimeConfigFromCLI(c))
				return nil
			},
		},
		{
			Name:    "render",
			Aliases: []string{"r"},
			Usage:   "render gossfile after imports",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "debug, d",
					Usage: fmt.Sprintf("Print debugging info when rendering"),
				},
			},
			Action: func(c *cli.Context) error {
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
			Action: func(c *cli.Context) error {
				return goss.AutoAddResources(c.GlobalString("gossfile"), c.Args(), newRuntimeConfigFromCLI(c))
			},
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add a resource to the test suite",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "exclude-attr",
					Usage: "Exclude the following attributes when adding a new resource",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:  "package",
					Usage: "add new package",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Package", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "file",
					Usage: "add new file",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "File", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "addr",
					Usage: "add new remote address:port - ex: google.com:80",
					Flags: []cli.Flag{
						cli.DurationFlag{
							Name:  "timeout",
							Value: 500 * time.Millisecond,
						},
					},
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Addr", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "port",
					Usage: "add new listening [protocol]:port - ex: 80 or udp:123",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Port", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "service",
					Usage: "add new service",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Service", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "user",
					Usage: "add new user",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "User", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "group",
					Usage: "add new group",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Group", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "command",
					Usage: "add new command",
					Flags: []cli.Flag{
						cli.DurationFlag{
							Name:  "timeout",
							Value: 10 * time.Second,
						},
					},
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Command", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "dns",
					Usage: "add new dns",
					Flags: []cli.Flag{
						cli.DurationFlag{
							Name:  "timeout",
							Value: 500 * time.Millisecond,
						},
						cli.StringFlag{
							Name:  "server",
							Usage: "The IP address of a DNS server to query",
						},
					},
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "DNS", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "process",
					Usage: "add new process name",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Process", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "http",
					Usage: "add new http",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: "insecure, k",
						},
						cli.BoolFlag{
							Name: "no-follow-redirects, r",
						},
						cli.DurationFlag{
							Name:  "timeout",
							Value: 5 * time.Second,
						},
						cli.StringFlag{
							Name:  "username, u",
							Usage: "Username for basic auth",
						},
						cli.StringFlag{
							Name:  "password, p",
							Usage: "Password for basic auth",
						},
					},
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "HTTP", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "goss",
					Usage: "add new goss file, it will be imported from this one",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Gossfile", c.Args(), newRuntimeConfigFromCLI(c))

					},
				},
				{
					Name:  "kernel-param",
					Usage: "add new goss kernel param",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "KernelParam", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "mount",
					Usage: "add new mount",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Mount", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
				{
					Name:  "interface",
					Usage: "add new interface",
					Action: func(c *cli.Context) error {
						return goss.AddResources(c.GlobalString("gossfile"), "Interface", c.Args(), newRuntimeConfigFromCLI(c))
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
