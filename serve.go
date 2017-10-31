package goss

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aelsabbahy/goss/outputs"
	"github.com/aelsabbahy/goss/system"
	"github.com/fatih/color"
	"github.com/patrickmn/go-cache"
	"github.com/urfave/cli"
)

func Serve(c *cli.Context) {
	endpoint := c.String("endpoint")
	color.NoColor = true
	cache := cache.New(c.Duration("cache"), 30*time.Second)
	o, err := getOutputer(c)
	if err != nil {
		log.Fatal(err)
	}

	health := healthHandler{
		c:             c,
		gossConfig:    getGossConfig(c),
		sys:           system.New(c),
		outputer:      o,
		cache:         cache,
		gossMu:        &sync.Mutex{},
		maxConcurrent: c.Int("max-concurrent"),
	}
	if c.String("format") == "json" {
		health.contentType = "application/json"
	}
	http.Handle(endpoint, health)
	listenAddr := c.String("listen-addr")
	log.Printf("Starting to listen on: %s", listenAddr)
	log.Fatal(http.ListenAndServe(c.String("listen-addr"), nil))
}

type res struct {
	exitCode int
	b        bytes.Buffer
}
type healthHandler struct {
	c             *cli.Context
	gossConfig    GossConfig
	sys           *system.System
	outputer      outputs.Outputer
	cache         *cache.Cache
	gossMu        *sync.Mutex
	contentType   string
	maxConcurrent int
}

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			h.sys = system.New(h.c)
			log.Printf("%v: Stale cache, running tests", r.RemoteAddr)
			iStartTime := time.Now()
			out := validate(h.sys, h.gossConfig, h.maxConcurrent)
			var b bytes.Buffer
			exitCode := h.outputer.Output(&b, out, iStartTime)
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
