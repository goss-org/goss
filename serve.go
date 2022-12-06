package goss

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/goss-org/goss/outputs"
	"github.com/goss-org/goss/resource"
	"github.com/goss-org/goss/system"
	"github.com/goss-org/goss/util"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Serve(c *util.Config) error {
	endpoint := c.Endpoint
	health, err := newHealthHandler(c)
	if err != nil {
		return err
	}
	http.Handle(endpoint, health)
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting to listen on: %s", c.ListenAddress)
	return http.ListenAndServe(c.ListenAddress, nil)
}

func newHealthHandler(c *util.Config) (*healthHandler, error) {
	color.NoColor = true
	cache := cache.New(c.Cache, 30*time.Second)

	cfg, err := getGossConfig(c.Vars, c.VarsInline, c.Spec)
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

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	outputFormat, outputer, err := h.negotiateResponseContentType(r)
	if err != nil {
		log.Printf("Warn: Using process-level output-format. %s", err)
		outputFormat = h.c.OutputFormat
		outputer = h.outputer
	}
	negotiatedContentType := h.responseContentType(outputFormat)

	log.Printf("%v: requesting health probe", r.RemoteAddr)
	resp := h.processAndEnsureCached(negotiatedContentType, outputer)
	w.Header().Set(http.CanonicalHeaderKey("Content-Type"), negotiatedContentType)
	w.WriteHeader(resp.statusCode)
	logBody := ""
	if resp.statusCode != http.StatusOK {
		logBody = " - " + resp.body.String()
	}
	resp.body.WriteTo(w)
	log.Printf("%v: status %d%s", r.RemoteAddr, resp.statusCode, logBody)
}

func (h healthHandler) processAndEnsureCached(negotiatedContentType string, outputer outputs.Outputer) res {
	cacheKey := fmt.Sprintf("res:%s", negotiatedContentType)
	tmp, found := h.cache.Get(cacheKey)
	if found {
		return tmp.(res)
	}

	h.gossMu.Lock()
	defer h.gossMu.Unlock()
	tmp, found = h.cache.Get(cacheKey)
	if found {
		log.Printf("Returning cached[%s].", cacheKey)
		return tmp.(res)
	}

	log.Printf("Stale cache[%s], running tests", cacheKey)
	resp := h.runValidate(outputer)
	h.cache.SetDefault(cacheKey, resp)
	return resp
}

func (h healthHandler) runValidate(outputer outputs.Outputer) res {
	h.sys = system.New(h.c.PackageManager)
	iStartTime := time.Now()
	out := validate(h.sys, h.gossConfig, h.c.DisabledResourceTypes, h.maxConcurrent)
	var b bytes.Buffer
	outputConfig := util.OutputConfig{
		FormatOptions: h.c.FormatOptions,
	}
	exitCode := outputer.Output(&b, out, iStartTime, outputConfig)
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
	return fmt.Sprintf("%s%s", mediaTypePrefix, outputName)
}

func (h healthHandler) renderBody(results <-chan []resource.TestResult, outputer outputs.Outputer) (int, bytes.Buffer) {
	outputConfig := util.OutputConfig{
		FormatOptions: h.c.FormatOptions,
	}
	var b bytes.Buffer
	exitCode := outputer.Output(&b, results, time.Now(), outputConfig)
	return exitCode, b
}
