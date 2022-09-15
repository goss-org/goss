package system

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/util"
)

type HTTP interface {
	HTTP() string
	Status() (int, error)
	Headers() ([]string, error)
	Body() (io.Reader, error)
	Exists() (bool, error)
	SetAllowInsecure(bool)
	SetNoFollowRedirects(bool)
}

type DefHTTP struct {
	http              string
	allowInsecure     bool
	noFollowRedirects bool
	resp              *http.Response
	RequestHeader     http.Header
	RequestBody       string
	Timeout           int
	loaded            bool
	err               error
	Username          string
	Password          string
	Method            string
	Proxy             string
}

func NewDefHTTP(httpStr string, system *System, config util.Config) HTTP {
	headers := http.Header{}
	for _, r := range config.RequestHeader {
		str := strings.SplitN(r, ": ", 2)
		headers.Add(str[0], str[1])
	}
	return &DefHTTP{
		http:              httpStr,
		allowInsecure:     config.AllowInsecure,
		Method:            config.Method,
		noFollowRedirects: config.NoFollowRedirects,
		RequestHeader:     headers,
		RequestBody:       config.RequestBody,
		Timeout:           config.TimeOutMilliSeconds(),
		Username:          config.Username,
		Password:          config.Password,
		Proxy:             config.Proxy,
	}
}

func HeaderToArray(header http.Header) (res []string) {
	for name, values := range header {
		for _, value := range values {
			res = append(res, strings.ToLower(fmt.Sprintf("%s: %s", name, value)))
		}
	}
	sort.Strings(res)
	return
}

func (u *DefHTTP) setup() error {
	if u.loaded {
		return u.err
	}
	u.loaded = true

	proxyURL := http.ProxyFromEnvironment
	if u.Proxy != "" {
		parseProxy, err := url.Parse(u.Proxy)

		if err != nil {
			return err
		}

		proxyURL = http.ProxyURL(parseProxy)
	}

	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: u.allowInsecure},
		DisableKeepAlives: true,
		Proxy:             proxyURL,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(u.Timeout) * time.Millisecond,
	}

	if u.noFollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest(u.Method, u.http, strings.NewReader(u.RequestBody))
	if err != nil {
		return u.err
	}
	req.Header = u.RequestHeader.Clone()

	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	if u.Username != "" || u.Password != "" {
		req.SetBasicAuth(u.Username, u.Password)
	}
	u.resp, u.err = client.Do(req)

	return u.err
}

func (u *DefHTTP) Exists() (bool, error) {
	if _, err := u.Status(); err != nil {
		return false, err
	}
	return true, nil
}

func (u *DefHTTP) SetNoFollowRedirects(t bool) {
	u.noFollowRedirects = t
}

func (u *DefHTTP) SetAllowInsecure(t bool) {
	u.allowInsecure = t
}

func (u *DefHTTP) ID() string {
	return u.http
}

func (u *DefHTTP) HTTP() string {
	return u.http
}

func (u *DefHTTP) Status() (int, error) {
	if err := u.setup(); err != nil {
		return 0, err
	}

	return u.resp.StatusCode, nil
}

func (u *DefHTTP) Headers() ([]string, error) {
	if err := u.setup(); err != nil {
		return nil, err
	}

	return HeaderToArray(u.resp.Header), nil
}

func (u *DefHTTP) Body() (io.Reader, error) {
	if err := u.setup(); err != nil {
		return nil, err
	}

	return u.resp.Body, nil
}
