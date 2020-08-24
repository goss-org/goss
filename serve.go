package goss

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
	"github.com/fatih/color"
	"github.com/patrickmn/go-cache"
)

func Serve(c *util.Config) {
	endpoint := c.Endpoint
	health, err := newHealthHandler(c)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	http.Handle(endpoint, health)
	log.Printf("Starting to listen on: %s", c.ListenAddress)
	log.Fatal(http.ListenAndServe(c.ListenAddress, nil))
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
	if c.OutputFormat == "json" {
		health.contentType = "application/json"
	}
	return health, nil
}

type res struct {
	exitCode int
	b        bytes.Buffer
}
type healthHandler struct {
	c             *util.Config
	gossConfig    GossConfig
	sys           *system.System
	outputer      outputs.Outputer
	cache         *cache.Cache
	gossMu        *sync.Mutex
	contentType   string
	maxConcurrent int
}

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	outputConfig := util.OutputConfig{
		FormatOptions: h.c.FormatOptions,
	}

	log.Printf("%v: requesting health probe", r.RemoteAddr)
	var resp res
	tmp, found := h.cache.Get("res")
	if found {
		resp = tmp.(res)
	} else {
		h.gossMu.Lock()
		defer h.gossMu.Unlock()
		tmp, found := h.cache.Get("res")
		if found {
			resp = tmp.(res)
		} else {
			h.sys = system.New(h.c.PackageManager)
			log.Printf("%v: Stale cache, running tests", r.RemoteAddr)
			iStartTime := time.Now()
			out := validate(h.sys, h.gossConfig, h.maxConcurrent)
			var b bytes.Buffer
			exitCode := h.outputer.Output(&b, out, iStartTime, outputConfig)
			resp = res{exitCode: exitCode, b: b}
			h.cache.Set("res", resp, cache.DefaultExpiration)
		}
	}
	if h.contentType != "" {
		w.Header().Set("Content-Type", h.contentType)
	}
	if resp.exitCode == 0 {
		resp.b.WriteTo(w)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		resp.b.WriteTo(w)
	}
}
