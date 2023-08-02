package system

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/goss-org/goss/util"
)

type HTTP interface {
	HTTP() string
	Status() (int, error)
	Headers() (io.Reader, error)
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
	CAFile            string
	CertFile          string
	KeyFile           string
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
		CAFile:            config.CAFile,
		CertFile:          config.CertFile,
		KeyFile:           config.KeyFile,
		Proxy:             config.Proxy,
	}
}

func HeaderToArray(header http.Header) (res []string) {
	for name, values := range header {
		for _, value := range values {
			res = append(res, fmt.Sprintf("%s: %s", name, value))
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

	tlsConfig := &tls.Config{
		InsecureSkipVerify: u.allowInsecure,
		Renegotiation:      tls.RenegotiateFreelyAsClient,
	}
	if u.CAFile != "" {
		// FIXME: iotutil
		caCert, err := os.ReadFile(u.CAFile)
		if err != nil {
			return err
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(caCert)
		if !ok {
			return fmt.Errorf("Failed parse root certificate: %s", u.CAFile)
		}
		tlsConfig.RootCAs = roots
	}

	if u.CertFile != "" && u.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(u.CertFile, u.KeyFile)
		if err != nil {
			return err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	tr := &http.Transport{
		TLSClientConfig:   tlsConfig,
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

func (u *DefHTTP) Headers() (io.Reader, error) {
	if err := u.setup(); err != nil {
		return nil, err
	}

	var headerString = strings.Join(HeaderToArray(u.resp.Header), "\n")
	return strings.NewReader(headerString), nil
}

func (u *DefHTTP) Body() (io.Reader, error) {
	if err := u.setup(); err != nil {
		return nil, err
	}

	return u.resp.Body, nil
}
