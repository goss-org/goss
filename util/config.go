package util

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/oleiade/reflections"
)

// ConfigOption manipulates Config
type ConfigOption func(c *Config) error

// Config is the runtime configuration for the goss system, the cli.Context gets
// converted to this and it allows other packages to embed goss by creating this
// structure and using it when adding, validating etc.
//
// NewConfig can be used to create this which will default to what the CLI assumes
// and allow manipulation via ConfigOption functions
type Config struct {
	AllowInsecure     bool
	AnnounceToCLI     bool
	Cache             time.Duration
	Debug             bool
	Endpoint          string
	FormatOptions     []string
	IgnoreList        []string
	ListenAddress     string
	LocalAddress      string
	MaxConcurrent     int
	NoColor           *bool
	NoFollowRedirects bool
	OutputFormat      string
	OutputWriter      io.Writer
	PackageManager    string
	Password          string
	RequestHeader     []string
	RetryTimeout      time.Duration
	Server            string
	Sleep             time.Duration
	Spec              string
	Timeout           time.Duration
	Username          string
	Vars              string
	VarsInline        string
}

// TimeOutMilliSeconds is the timeout as milliseconds
func (c *Config) TimeOutMilliSeconds() int {
	return int(c.Timeout / time.Millisecond)
}

// NewConfig creates a default configuration modeled on the defaults the CLI sets, modified using opts
func NewConfig(opts ...ConfigOption) (rc *Config, err error) {
	rc = &Config{
		AllowInsecure:     false,
		AnnounceToCLI:     false,
		Cache:             5 * time.Second,
		Debug:             false,
		Endpoint:          "/healthz",
		FormatOptions:     []string{},
		IgnoreList:        []string{},
		ListenAddress:     ":8080",
		LocalAddress:      "",
		MaxConcurrent:     50,
		NoColor:           nil,
		NoFollowRedirects: false,
		OutputFormat:      "structured", // most appropriate for package usage
		PackageManager:    "",
		Password:          "",
		RequestHeader:     nil,
		RetryTimeout:      0,
		Server:            "",
		Sleep:             time.Second,
		Spec:              "",
		Timeout:           0,
		Username:          "",
		Vars:              "",
		VarsInline:        "",
	}

	// NewConfig() is likely to be used when embedding goss or using as a package
	// so assuming no color seems like a sane departure from CLI defaults
	WithNoColor()(rc)

	for _, opt := range opts {
		err = opt(rc)
		if err != nil {
			return nil, err
		}
	}

	return rc, nil
}

// WithSpecFile sets the path to the file holding spec contents
func WithSpecFile(f string) ConfigOption {
	return func(c *Config) error {
		c.Spec = f
		return nil
	}
}

// WithOutputFormat is the formatter to use for output
func WithOutputFormat(f string) ConfigOption {
	return func(c *Config) error {
		c.OutputFormat = f

		return nil
	}
}

// WithFormatOptions sets options used by the output format plugins, valid options are output.WithFormatOptions
func WithFormatOptions(opts ...string) ConfigOption {
	return func(c *Config) error {
		for _, o := range opts {
			c.FormatOptions = append(c.FormatOptions, o)
		}

		return nil
	}
}

// WithResultWriter sets the writer to write output format to when validating
func WithResultWriter(w io.Writer) ConfigOption {
	return func(c *Config) error {
		c.OutputWriter = w
		return nil
	}
}

// WithSleep sets the time to sleep between retries when WithRetryTimeout is set
func WithSleep(d time.Duration) ConfigOption {
	return func(c *Config) error {
		c.Sleep = d
		return nil
	}
}

// WithRetryTimeout sets the maximum amount of time checks can be retried, it's runtime + WithSleep
func WithRetryTimeout(d time.Duration) ConfigOption {
	return func(c *Config) error {
		c.RetryTimeout = d
		return nil
	}
}

// WithCache sets how long results may be cached for
func WithCache(d time.Duration) ConfigOption {
	return func(c *Config) error {
		c.Cache = d
		return nil
	}
}

// WithMaxConcurrency is the maximum concurrent test that can be run
func WithMaxConcurrency(mc int) ConfigOption {
	return func(c *Config) error {
		c.MaxConcurrent = mc
		return nil
	}
}

// WithNoColor disables colored output
func WithNoColor() ConfigOption {
	return func(c *Config) error {
		c.NoColor = func(b bool) *bool { return &b }(true)
		return nil
	}
}

// WithColor enables colored output
func WithColor() ConfigOption {
	return func(c *Config) error {
		c.NoColor = func(b bool) *bool { return &b }(false)
		return nil
	}
}

// WithPackageManager overrides the package manager to use
func WithPackageManager(p string) ConfigOption {
	return func(c *Config) error {
		c.PackageManager = p

		return nil
	}
}

// WithDebug enables debug output
func WithDebug() ConfigOption {
	return func(c *Config) error {
		c.Debug = true
		return nil
	}
}

// WithVarsFile is a json or yaml file containing variables to pass to the validator
func WithVarsFile(file string) ConfigOption {
	return func(c *Config) error {
		c.Vars = file
		return nil
	}
}

// WithVarsData uses v as variables to pass to the Validator
func WithVarsData(v interface{}) ConfigOption {
	return func(c *Config) error {
		jv, err := json.Marshal(v)
		if err != nil {
			return err
		}

		c.VarsInline = string(jv)

		return nil
	}
}

// WithVarsBytes is a yaml or json byte stream to use as variables passed to the Validator
func WithVarsBytes(v []byte) ConfigOption {
	return WithVarsString(string(v))
}

// WithVarsString is a yaml or json string to use as variables passed to the Validator
func WithVarsString(v string) ConfigOption {
	return func(c *Config) error {
		c.VarsInline = v
		return nil
	}
}

type OutputConfig struct {
	FormatOptions []string
}

type format string

const (
	JSON format = "json"
	YAML format = "yaml"
)

func ValidateSections(unmarshal func(interface{}) error, i interface{}, whitelist map[string]bool) error {
	// Get generic input
	var toValidate map[string]map[string]interface{}
	if err := unmarshal(&toValidate); err != nil {
		return err
	}

	// Run input through whitelist
	typ := reflect.TypeOf(i)
	typs := strings.Split(typ.String(), ".")[1]
	for id, v := range toValidate {
		for k, _ := range v {
			if !whitelist[k] {
				return fmt.Errorf("invalid Attribute for %s:%s: %s", typs, id, k)
			}
		}
	}

	return nil
}

func WhitelistAttrs(i interface{}, format format) (map[string]bool, error) {
	validAttrs := make(map[string]bool)
	tags, err := reflections.Tags(i, string(format))
	if err != nil {
		return nil, err
	}
	for _, v := range tags {
		validAttrs[strings.Split(v, ",")[0]] = true
	}
	return validAttrs, nil
}

func IsValueInList(value string, list []string) bool {
	for _, v := range list {
		if strings.ToLower(v) == strings.ToLower(value) {
			return true
		}
	}
	return false
}
