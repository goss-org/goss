package goss

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goss-org/goss/outputs"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Serve(c *util.Config) error {
	err := setLogLevel(c)
	if err != nil {
		return err
	}
	logger := util.LoggerOrDiscard(c.Logger)
	endpoint := c.Endpoint
	health, err := newHealthHandler(c)
	if err != nil {
		return err
	}
	http.Handle(endpoint, health)
	http.Handle("/metrics", promhttp.Handler())
	logger.Info("server listening", "listen_addr", c.ListenAddress)
	return newServer(c, nil).ListenAndServe()
}

// newServer builds the server Serve runs.
//
// Production passes a nil handler, which is what http.ListenAndServe does and
// keeps the endpoints on http.DefaultServeMux exactly as before. The parameter
// exists so that a test can drive a real server without registering on a
// process-global mux, which can only be done once per process.
func newServer(c *util.Config, h http.Handler) *http.Server {
	logger := util.LoggerOrDiscard(c.Logger)

	return &http.Server{
		Addr:    c.ListenAddress,
		Handler: h,
		// http.Server reports through a *log.Logger, so what arrives here is
		// one preformatted message with no structure to recover. Routing it at
		// ERROR through the configured handler is the whole of what this
		// bridge can do, and is the reason the message is the only goss record
		// that is not built from attributes.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
}

func newHealthHandler(c *util.Config) (*healthHandler, error) {
	// The health endpoint always emits machine-readable output.
	outputs.SetNoColor(true)
	cache := cache.New(c.Cache, 30*time.Second)

	cfg, err := getGossConfig(c)
	if err != nil {
		return nil, err
	}

	output, err := getOutputer(c.NoColor, c.OutputFormat)
	if err != nil {
		return nil, err
	}

	logger := util.LoggerOrDiscard(c.Logger)

	health := &healthHandler{
		c:             c,
		logger:        logger,
		gossConfig:    *cfg,
		sys:           system.New(c.PackageManager, system.WithLogger(logger)),
		outputer:      output,
		cache:         cache,
		gossMu:        &sync.Mutex{},
		maxConcurrent: c.MaxConcurrent,
	}
	return health, nil
}

type res struct {
	body       bytes.Buffer
	statusCode int
}
type healthHandler struct {
	c             *util.Config
	logger        *slog.Logger
	gossConfig    GossConfig
	sys           *system.System
	outputer      outputs.Outputer
	cache         *cache.Cache
	gossMu        *sync.Mutex
	maxConcurrent int
}

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	outputFormat, outputer, err := h.negotiateResponseContentType(r)
	if err != nil {
		h.logger.Debug("using configured output format", slog.Any("error", err))
		outputFormat = h.c.OutputFormat
		outputer = h.outputer
	}
	negotiatedContentType := h.responseContentType(outputFormat)

	util.TraceContext(r.Context(), h.logger, "request received", "client_addr", r.RemoteAddr)
	resp := h.processAndEnsureCached(r.Context(), negotiatedContentType, outputer)
	w.Header().Set("Content-Type", negotiatedContentType)
	w.WriteHeader(resp.statusCode)
	// The response body is only worth logging when something went wrong, and
	// keeping that gate matters: it is what stops a successful probe writing
	// every check's result into the log on every request.
	attrs := []any{"client_addr", r.RemoteAddr, "http_status", resp.statusCode}
	if resp.statusCode != http.StatusOK {
		attrs = append(attrs, "response_body", resp.body.String())
	}
	resp.body.WriteTo(w)
	h.logger.Debug("request complete", attrs...)
}

func (h healthHandler) processAndEnsureCached(ctx context.Context, negotiatedContentType string, outputer outputs.Outputer) res {
	var tra [][]resource.TestResult
	cacheKey := "res"
	// Held across the lookup so a miss does not let a second request start its
	// own validate() before the first one has filled the cache.
	h.gossMu.Lock()
	if tmp, found := h.cache.Get(cacheKey); found {
		util.TraceContext(ctx, h.logger, "returning cached result", "cache_key", cacheKey)
		tra = tmp.([][]resource.TestResult)
	} else {
		util.TraceContext(ctx, h.logger, "running validation for stale cache", "cache_key", cacheKey)
		tra = h.validate()
		h.cache.SetDefault(cacheKey, tra)
	}
	h.gossMu.Unlock()

	trc := testResultArrayToChan(tra)
	return h.output(trc, outputer)
}

func (h healthHandler) output(trc <-chan []resource.TestResult, outputer outputs.Outputer) res {
	var b bytes.Buffer
	outputConfig := util.OutputConfig{
		FormatOptions: h.c.FormatOptions,
		Logger:        h.logger,
	}
	exitCode := outputer.Output(&b, trc, outputConfig)
	resp := res{
		body: b,
	}
	if exitCode == 0 {
		resp.statusCode = http.StatusOK
	} else {
		resp.statusCode = http.StatusServiceUnavailable
	}
	return resp
}
func (h healthHandler) validate() [][]resource.TestResult {
	h.sys = system.New(h.c.PackageManager, system.WithLogger(h.logger))
	res := make([][]resource.TestResult, 0)
	tr := validate(h.sys, h.gossConfig, h.c.DisabledResourceTypes, h.maxConcurrent)
	for i := range tr {
		res = append(res, i)
	}
	return res
}

func testResultArrayToChan(tra [][]resource.TestResult) <-chan []resource.TestResult {
	c := make(chan []resource.TestResult)
	go func(c chan []resource.TestResult) {
		defer close(c)

		for _, i := range tra {
			c <- i
		}
	}(c)

	return c
}

const (
	// https://en.wikipedia.org/wiki/Media_type
	mediaTypePrefix = "application/vnd.goss-"
)

func (h healthHandler) negotiateResponseContentType(r *http.Request) (string, outputs.Outputer, error) {
	acceptHeader := r.Header[http.CanonicalHeaderKey("Accept")]
	var outputer outputs.Outputer
	outputName := ""
	for _, acceptCandidate := range acceptHeader {
		acceptCandidate = strings.TrimSpace(acceptCandidate)
		if strings.HasPrefix(acceptCandidate, mediaTypePrefix) {
			outputName = strings.TrimPrefix(acceptCandidate, mediaTypePrefix)
		} else if strings.EqualFold("application/json", acceptCandidate) || strings.EqualFold("text/json", acceptCandidate) {
			outputName = "json"
		} else {
			outputName = ""
		}
		var err error
		outputer, err = outputs.GetOutputer(outputName)
		if err != nil {
			continue
		}
	}
	if outputer == nil {
		return "", nil, fmt.Errorf("Accept header on request missing or invalid. Accept header: %v", acceptHeader)
	}

	return outputName, outputer, nil
}

func (h healthHandler) responseContentType(outputName string) string {
	if outputName == "json" {
		return "application/json"
	}
	if outputName == "prometheus" {
		return "text/plain; version=0.0.4"
	}

	return fmt.Sprintf("%s%s", mediaTypePrefix, outputName)
}
