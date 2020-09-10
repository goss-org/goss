package resource

import (
	"time"

	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
)

type HTTP struct {
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Meta              meta     `json:"meta,omitempty" yaml:"meta,omitempty"`
	HTTP              string   `json:"-" yaml:"-"`
	Status            matcher  `json:"status" yaml:"status"`
	AllowInsecure     bool     `json:"allow-insecure" yaml:"allow-insecure"`
	NoFollowRedirects bool     `json:"no-follow-redirects" yaml:"no-follow-redirects"`
	Timeout           int      `json:"timeout" yaml:"timeout"`
	RequestHeader     []string `json:"request-headers,omitempty" yaml:"request-headers,omitempty"`
	Headers           []string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body              []string `json:"body" yaml:"body"`
	Username          string   `json:"username,omitempty" yaml:"username,omitempty"`
	Password          string   `json:"password,omitempty" yaml:"password,omitempty"`
	Skip              bool     `json:"skip,omitempty" yaml:"skip,omitempty"`
}

func (u *HTTP) ID() string      { return u.HTTP }
func (u *HTTP) SetID(id string) { u.HTTP = id }

// FIXME: Can this be refactored?
func (r *HTTP) GetTitle() string { return r.Title }
func (r *HTTP) GetMeta() meta    { return r.Meta }

func (u *HTTP) Validate(sys *system.System) []TestResult {
	skip := false
	if u.Timeout == 0 {
		u.Timeout = 5000
	}
	sysHTTP := sys.NewHTTP(u.HTTP, sys, util.Config{
		AllowInsecure: u.AllowInsecure, NoFollowRedirects: u.NoFollowRedirects,
		Timeout: time.Duration(u.Timeout) * time.Millisecond, Username: u.Username, Password: u.Password,
		RequestHeader: u.RequestHeader})
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
	}
	return u, err
}
