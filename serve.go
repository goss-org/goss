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
	resp.body.WriteTo(w)
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
	out := validate(h.sys, h.gossConfig, h.maxConcurrent)
	var b bytes.Buffer
	outputConfig := util.OutputConfig{
		FormatOptions: h.c.FormatOptions,
	}
	exitCode := outputer.Output(&b, out, iStartTime, outputConfig)
	resp := res{
		body: b,
	}
	log.Printf("exitCode: %d", exitCode)
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
	found := r.Header[http.CanonicalHeaderKey("Accept")]
	var outputer outputs.Outputer
	outputName := ""
	for _, header := range found {
		candidates := strings.Split(header, ",")
		for _, acceptCandidate := range candidates {
			acceptCandidate = strings.TrimSpace(acceptCandidate)
			log.Printf("candidate %q", acceptCandidate)
			if strings.HasPrefix(acceptCandidate, mediaTypePrefix) {
				outputName = strings.TrimLeft(acceptCandidate, mediaTypePrefix)
			} else if strings.EqualFold("application/json", acceptCandidate) || strings.EqualFold("text/json", acceptCandidate) {
				outputName = "json"
			}
			var err error
			outputer, err = outputs.GetOutputer(outputName)
			if err != nil {
				continue
			}
		}
	}
	if outputer == nil {
		return "", nil, fmt.Errorf("Accept header on request missing or invalid. Accept header: %v", found)
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
