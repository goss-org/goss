package goss

import (
	"bytes"
	"fmt"
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
	endpoint := c.Endpoint
	health, err := newHealthHandler(c)
	if err != nil {
		return err
	}
	http.Handle(endpoint, health)
	http.Handle("/metrics", promhttp.Handler())
	c.Log().Printf("[INFO] Starting to listen on: %s", c.ListenAddress)
	return http.ListenAndServe(c.ListenAddress, nil)
}

func newHealthHandler(c *util.Config) (*healthHandler, error) {
	// The serve endpoint always produces machine-readable output, so disable
	// ANSI color codes. Using util.InitNoColor (sync.Once) avoids racing with
	// concurrent requests or other initialization paths.
	util.InitNoColor(true)
	cache := cache.New(c.Cache, 30*time.Second)

	cfg, err := getGossConfig(c, c.Vars, c.VarsInline, c.Spec)
	if err != nil {
		return nil, err
	}

	output, err := getOutputer(c.NoColor, c.OutputFormat)
	if err != nil {
		return nil, err
	}

	health := &healthHandler{
		c:             c,
		gossConfig:    *cfg,
		sys:           system.New(c.PackageManager),
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
	gossConfig    GossConfig
	sys           *system.System
	outputer      outputs.Outputer
	cache         *cache.Cache
	gossMu        *sync.Mutex
	maxConcurrent int
}

// markFilter holds the mark filters effective for a single request. Query
// parameters take precedence over the values from the server config.
type markFilter struct {
	includeMarks []string
	excludeMarks []string
}

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.c.Log()
	outputFormat, outputer, err := h.negotiateResponseContentType(r)
	if err != nil {
		logger.Printf("[DEBUG] Warn: Using process-level output-format. %s", err)
		outputFormat = h.c.OutputFormat
		outputer = h.outputer
	}
	negotiatedContentType := h.responseContentType(outputFormat)

	mf := h.resolveMarkFilter(r)

	logger.Printf("[TRACE] %v: requesting health probe", r.RemoteAddr)
	resp := h.processAndEnsureCached(negotiatedContentType, outputer, mf)
	w.Header().Set(http.CanonicalHeaderKey("Content-Type"), negotiatedContentType) //nolint:gosimple
	w.WriteHeader(resp.statusCode)
	logBody := ""
	if resp.statusCode != http.StatusOK {
		logBody = " - " + resp.body.String()
	}
	resp.body.WriteTo(w)
	logger.Printf("[DEBUG] %v: status %d%s", r.RemoteAddr, resp.statusCode, logBody)
}

// resolveMarkFilter builds the effective mark filter for a request.
// Query parameters override the server-level config values when set.
func (h healthHandler) resolveMarkFilter(r *http.Request) markFilter {
	mf := markFilter{
		includeMarks: h.c.IncludeMarks,
		excludeMarks: h.c.ExcludeMarks,
	}
	q := r.URL.Query()
	if q.Has("marks") {
		mf.includeMarks = util.ParseMarksParam(q.Get("marks"))
	}
	if q.Has("exclude-marks") {
		mf.excludeMarks = util.ParseMarksParam(q.Get("exclude-marks"))
	}
	return mf
}

// cacheKey returns a unique cache key per (include, exclude) mark combination.
// The empty filter case yields the legacy "res" key, preserving existing
// cache semantics for callers that do not use marks.
func (mf markFilter) cacheKey() string {
	if len(mf.includeMarks) == 0 && len(mf.excludeMarks) == 0 {
		return "res"
	}
	key := "res"
	if len(mf.includeMarks) > 0 {
		key += ":include=" + strings.Join(mf.includeMarks, ",")
	}
	if len(mf.excludeMarks) > 0 {
		key += ":exclude=" + strings.Join(mf.excludeMarks, ",")
	}
	return key
}

func (h healthHandler) processAndEnsureCached(negotiatedContentType string, outputer outputs.Outputer, mf markFilter) res {
	logger := h.c.Log()
	var tra [][]resource.TestResult
	cacheKey := mf.cacheKey()
	tmp, found := h.cache.Get(cacheKey)
	if found {
		logger.Printf("[TRACE] Returning cached[%s].", cacheKey)
		tra = tmp.([][]resource.TestResult)
	} else {
		logger.Printf("Stale cache[%s], running tests", cacheKey)
		h.sys = system.New(h.c.PackageManager)
		tra = h.validate(mf)
		h.cache.SetDefault(cacheKey, tra)
	}
	trc := testResultArrayToChan(tra)
	return h.output(trc, outputer)
}

func (h healthHandler) output(trc <-chan []resource.TestResult, outputer outputs.Outputer) res {
	var b bytes.Buffer
	outputConfig := util.OutputConfig{
		FormatOptions: h.c.FormatOptions,
		Logger:        h.c.Logger,
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
func (h healthHandler) validate(mf markFilter) [][]resource.TestResult {
	// Serialize concurrent validations against the shared gossConfig so that
	// per-request mark filters do not leak Skip state across requests.
	// validate() may call SetSkip() on resources to filter them; we snapshot
	// and restore the originals to keep the in-memory config stable.
	h.gossMu.Lock()
	defer h.gossMu.Unlock()

	originalSkips := snapshotSkips(h.gossConfig)
	defer restoreSkips(h.gossConfig, originalSkips)

	h.sys = system.New(h.c.PackageManager)
	res := make([][]resource.TestResult, 0)
	tr := validate(h.sys, h.gossConfig, h.c.DisabledResourceTypes, mf.includeMarks, mf.excludeMarks, h.maxConcurrent, h.c.Log())
	for i := range tr {
		res = append(res, i)
	}
	return res
}

// snapshotSkips records the current Skip state of every resource so it can be
// restored after validate() mutates them via SetSkip().
//
// Background: validate() calls Resource.SetSkip() to filter resources from
// execution.  In the serve path, the gossConfig is shared across requests, so
// per-request filters (e.g. ?marks=critical) would otherwise leak Skip state
// into subsequent requests.  We snapshot before validation and restore after
// to keep the in-memory config stable.
func snapshotSkips(cfg GossConfig) map[string]bool {
	m := make(map[string]bool)
	for _, r := range cfg.Resources() {
		m[resourceKey(r)] = readSkip(r)
	}
	return m
}

// restoreSkips re-applies the original Skip flags so subsequent validate()
// invocations start from a clean state.
func restoreSkips(cfg GossConfig, originals map[string]bool) {
	for _, r := range cfg.Resources() {
		key := resourceKey(r)
		writeSkip(r, originals[key])
	}
}

// resourceKey produces a stable identity for a Resource within a single
// gossConfig.  TypeKey + ID is unique because each resource type has its
// own keyed map in the config.
func resourceKey(r resource.Resource) string {
	type ider interface{ ID() string }
	id := ""
	if i, ok := r.(ider); ok {
		id = i.ID()
	}
	return r.TypeKey() + ":" + id
}

// readSkip returns the current Skip flag of a resource via a type switch.
// We avoid reflection for hot-path performance.
func readSkip(r resource.Resource) bool {
	switch v := r.(type) {
	case *resource.Addr:
		return v.Skip
	case *resource.Command:
		return v.Skip
	case *resource.DNS:
		return v.Skip
	case *resource.File:
		return v.Skip
	case *resource.Gossfile:
		return v.Skip
	case *resource.Group:
		return v.Skip
	case *resource.HTTP:
		return v.Skip
	case *resource.Interface:
		return v.Skip
	case *resource.KernelParam:
		return v.Skip
	case *resource.Matching:
		return v.Skip
	case *resource.Mount:
		return v.Skip
	case *resource.Package:
		return v.Skip
	case *resource.Port:
		return v.Skip
	case *resource.Process:
		return v.Skip
	case *resource.Service:
		return v.Skip
	case *resource.User:
		return v.Skip
	default:
		return false
	}
}

// writeSkip sets the Skip flag of a resource via a type switch.
func writeSkip(r resource.Resource, skip bool) {
	switch v := r.(type) {
	case *resource.Addr:
		v.Skip = skip
	case *resource.Command:
		v.Skip = skip
	case *resource.DNS:
		v.Skip = skip
	case *resource.File:
		v.Skip = skip
	case *resource.Gossfile:
		v.Skip = skip
	case *resource.Group:
		v.Skip = skip
	case *resource.HTTP:
		v.Skip = skip
	case *resource.Interface:
		v.Skip = skip
	case *resource.KernelParam:
		v.Skip = skip
	case *resource.Matching:
		v.Skip = skip
	case *resource.Mount:
		v.Skip = skip
	case *resource.Package:
		v.Skip = skip
	case *resource.Port:
		v.Skip = skip
	case *resource.Process:
		v.Skip = skip
	case *resource.Service:
		v.Skip = skip
	case *resource.User:
		v.Skip = skip
	}
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
