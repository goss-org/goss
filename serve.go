package goss

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/resource"
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
	"github.com/fatih/color"
	"github.com/patrickmn/go-cache"
)

func Serve(c *util.Config) error {
	endpoint := c.Endpoint
	health, err := newHealthHandler(c)
	if err != nil {
		return err
	}
	http.Handle(endpoint, health)
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

type runResult struct {
	body     bytes.Buffer
	exitCode int
}

func (rr runResult) toHTTPStatus() int {
	if rr.exitCode == 0 {
		return http.StatusOK
	}
	return http.StatusServiceUnavailable
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
	statusCode := resp.toHTTPStatus()
	w.WriteHeader(statusCode)
	logBody := ""
	if statusCode != http.StatusOK {
		// if there are any test failures, log all the details since machine state is volatile
		// and highly subject to intermittent failures compared to unit-tests.
		logBody = " - " + resp.body.String()
	}
	resp.body.WriteTo(w)
	log.Printf("%v: status %d%s", r.RemoteAddr, statusCode, logBody)
}

func (h healthHandler) processAndEnsureCached(negotiatedContentType string, outputer outputs.Outputer) runResult {
	result := h.runValidate(outputer)
	// Converted back to channel
	ch := make(chan []resource.TestResult)
	go func() {
		defer close(ch)
		for _, v := range result.testResults {
			ch <- v
		}
	}()
	return h.renderBody(ch, outputer, result.iStartTime)
}

// results is needed to track starttime and other metadata
type results struct {
	testResults [][]resource.TestResult
	iStartTime  time.Time
}

func (h healthHandler) runValidate(outputer outputs.Outputer) results {
	iStartTime := time.Now()
	tmp, found := h.cache.Get("res")
	if found {
		log.Printf("Returning cached.")
		return tmp.(results)
	}
	h.gossMu.Lock()
	defer h.gossMu.Unlock()
	log.Printf("Stale cache, running tests")
	h.sys = system.New(h.c.PackageManager)
	validateResult := validate(h.sys, h.gossConfig, h.maxConcurrent)
	var testResult [][]resource.TestResult // Cached as a slice
	for r := range validateResult {
		testResult = append(testResult, r)
	}
	res := results{testResults: testResult, iStartTime: iStartTime}
	h.cache.SetDefault("res", res)
	return res
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

func (h healthHandler) renderBody(results <-chan []resource.TestResult, outputer outputs.Outputer, startTime time.Time) runResult {
	outputConfig := util.OutputConfig{
		FormatOptions: h.c.FormatOptions,
	}
	var b bytes.Buffer
	exitCode := outputer.Output(&b, results, startTime, outputConfig)
	rr := runResult{
		body:     b,
		exitCode: exitCode,
	}
	return rr
}
