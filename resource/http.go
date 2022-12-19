package resource

import (
	"time"

	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
)

type HTTP struct {
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	URL               string   `json:"url,omitempty" yaml:"url,omitempty"`
	Meta              meta     `json:"meta,omitempty" yaml:"meta,omitempty"`
	HTTP              string   `json:"-" yaml:"-"`
	Method            string   `json:"method,omitempty" yaml:"method,omitempty"`
	Status            matcher  `json:"status" yaml:"status"`
	AllowInsecure     bool     `json:"allow-insecure" yaml:"allow-insecure"`
	NoFollowRedirects bool     `json:"no-follow-redirects" yaml:"no-follow-redirects"`
	Timeout           int      `json:"timeout" yaml:"timeout"`
	RequestHeader     []string `json:"request-headers,omitempty" yaml:"request-headers,omitempty"`
	RequestBody       string   `json:"request-body,omitemptyy" yaml:"request-body,omitempty"`
	Headers           []string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body              []string `json:"body" yaml:"body"`
	Username          string   `json:"username,omitempty" yaml:"username,omitempty"`
	Password          string   `json:"password,omitempty" yaml:"password,omitempty"`
	CAFile            string   `json:"ca-file,omitempty" yaml:"ca-file,omitempty"`
	CertFile          string   `json:"cert-file,omitempty" yaml:"cert-file,omitempty"`
	KeyFile           string   `json:"key-file,omitempty" yaml:"key-file,omitempty"`
	Skip              bool     `json:"skip,omitempty" yaml:"skip,omitempty"`
	Proxy             string   `json:"proxy,omitempty" yaml:"proxy,omitempty"`
}

const (
	HTTPResourceKey  = "http"
	HTTPResourceName = "HTTP"
)

func init() {
	registerResource(HTTPResourceKey, &HTTP{})
}

func (u *HTTP) ID() string { return u.HTTP }

func (u *HTTP) SetID(id string)  { u.HTTP = id }
func (u *HTTP) SetSkip()         { u.Skip = true }
func (u *HTTP) TypeKey() string  { return HTTPResourceKey }
func (u *HTTP) TypeName() string { return HTTPResourceName }

// FIXME: Can this be refactored?
func (u *HTTP) GetTitle() string { return u.Title }
func (u *HTTP) GetMeta() meta    { return u.Meta }

func (u *HTTP) getURL() string {
	if u.URL != "" {
		return u.URL
	}
	return u.HTTP
}

func (u *HTTP) Validate(sys *system.System) []TestResult {
	skip := u.Skip
	if u.Timeout == 0 {
		u.Timeout = 5000
	}
	sysHTTP := sys.NewHTTP(u.getURL(), sys, util.Config{
		AllowInsecure:     u.AllowInsecure,
		CAFile:            u.CAFile,
		CertFile:          u.CertFile,
		KeyFile:           u.KeyFile,
		NoFollowRedirects: u.NoFollowRedirects,
		Timeout:           time.Duration(u.Timeout) * time.Millisecond, Username: u.Username, Password: u.Password, Proxy: u.Proxy,
		RequestHeader: u.RequestHeader, RequestBody: u.RequestBody, Method: u.Method})
	sysHTTP.SetAllowInsecure(u.AllowInsecure)
	sysHTTP.SetNoFollowRedirects(u.NoFollowRedirects)

	var results []TestResult
	results = append(results, ValidateValue(u, "status", u.Status, sysHTTP.Status, skip))
	if shouldSkip(results) {
		skip = true
	}
	if len(u.Headers) > 0 {
		results = append(results, ValidateContains(u, "Headers", u.Headers, sysHTTP.Headers, skip))
	}
	if len(u.Body) > 0 {
		results = append(results, ValidateContains(u, "Body", u.Body, sysHTTP.Body, skip))
	}

	return results
}

func NewHTTP(sysHTTP system.HTTP, config util.Config) (*HTTP, error) {
	http := sysHTTP.HTTP()
	status, err := sysHTTP.Status()
	u := &HTTP{
		HTTP:              http,
		Status:            status,
		RequestHeader:     []string{},
		Headers:           []string{},
		Body:              []string{},
		AllowInsecure:     config.AllowInsecure,
		NoFollowRedirects: config.NoFollowRedirects,
		Timeout:           config.TimeOutMilliSeconds(),
		Username:          config.Username,
		Password:          config.Password,
		Proxy:             config.Proxy,
	}
	return u, err
}
