package resource

import (
	"time"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
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
	Timeout           int      `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	RequestHeader     []string `json:"request-headers,omitempty" yaml:"request-headers,omitempty"`
	RequestBody       string   `json:"request-bod,omitemptyy" yaml:"request-body,omitempty"`
	Headers           matcher  `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body              matcher  `json:"body,omitempty" yaml:"body,omitempty"`
	Username          string   `json:"username,omitempty" yaml:"username,omitempty"`
	Password          string   `json:"password,omitempty" yaml:"password,omitempty"`
	Skip              bool     `json:"skip,omitempty" yaml:"skip,omitempty"`
	Proxy             string   `json:"proxy,omitempty" yaml:"proxy,omitempty"`
}

func (u *HTTP) ID() string { return u.HTTP }

func (u *HTTP) SetID(id string) { u.HTTP = id }

// FIXME: Can this be refactored?
func (r *HTTP) GetTitle() string { return r.Title }
func (r *HTTP) GetMeta() meta    { return r.Meta }

func (r *HTTP) getURL() string {
	if r.URL != "" {
		return r.URL
	}
	return r.HTTP
}

func (u *HTTP) Validate(sys *system.System) []TestResult {
	skip := false
	if u.Timeout == 0 {
		u.Timeout = 5000
	}
	sysHTTP := sys.NewHTTP(u.getURL(), sys, util.Config{
		AllowInsecure: u.AllowInsecure, NoFollowRedirects: u.NoFollowRedirects,
		Timeout: time.Duration(u.Timeout) * time.Millisecond, Username: u.Username, Password: u.Password, Proxy: u.Proxy,
		RequestHeader: u.RequestHeader, RequestBody: u.RequestBody, Method: u.Method})
	sysHTTP.SetAllowInsecure(u.AllowInsecure)
	sysHTTP.SetNoFollowRedirects(u.NoFollowRedirects)

	if u.Skip {
		skip = true
	}

	var results []TestResult
	results = append(results, ValidateValue(u, "status", u.Status, sysHTTP.Status, skip))
	if shouldSkip(results) {
		skip = true
	}
	if isSet(u.Headers) {
		results = append(results, ValidateValue(u, "Headers", u.Headers, sysHTTP.Headers, skip))
	}
	if isSet(u.Body) {
		results = append(results, ValidateValue(u, "Body", u.Body, sysHTTP.Body, skip))
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
		Headers:           nil,
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
