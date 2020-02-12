package goss

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/system"
	"github.com/aelsabbahy/goss/util"
	"github.com/fatih/color"
	"github.com/patrickmn/go-cache"
)

func Serve(c *RuntimeConfig) {
	endpoint := c.Endpoint
	color.NoColor = true
	cache := cache.New(c.Cache, 30*time.Second)

	cfg, err := getGossConfig(c.Vars, c.VarsInline, c.Spec)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	output, err := getOutputer(c.NoColor, c.OutputFormat)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	health := healthHandler{
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
	http.Handle(endpoint, health)
	listenAddr := c.ListenAddress
	log.Printf("Starting to listen on: %s", listenAddr)
	log.Fatal(http.ListenAndServe(c.ListenAddress, nil))
}

type res struct {
	exitCode int
	b        bytes.Buffer
}
type healthHandler struct {
	c             *RuntimeConfig
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
