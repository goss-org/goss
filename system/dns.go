package system

import (
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/aelsabbahy/goss/util"
)

type DNS interface {
	Host() string
	Addrs() ([]string, error)
	Resolveable() (interface{}, error)
	Exists() (interface{}, error)
}

type DefDNS struct {
	host        string
	resolveable bool
	addrs       []string
	Timeout     int
	loaded      bool
	err         error
}

func NewDefDNS(host string, system *System, config util.Config) DNS {
	return &DefDNS{
		host:    host,
		Timeout: config.Timeout,
	}
}

func (d *DefDNS) Host() string {
	return d.host
}

func (d *DefDNS) setup() error {
	if d.loaded {
		return d.err
	}
	d.loaded = true

	addrs, err := lookupHost(d.host, d.Timeout)
	if err != nil || len(addrs) == 0 {
		d.resolveable = false
		d.addrs = []string{}
		// DNSError is resolvable == false, ignore error
		if _, ok := err.(*net.DNSError); ok {
			return nil
		}
		d.err = err
		return d.err
	}
	sort.Strings(addrs)
	d.resolveable = true
	d.addrs = addrs
	return nil
}

func (d *DefDNS) Addrs() ([]string, error) {
	err := d.setup()

	return d.addrs, err
}

func (d *DefDNS) Resolveable() (interface{}, error) {
	err := d.setup()

	return d.resolveable, err
}

// Stub out
func (d *DefDNS) Exists() (interface{}, error) {
	return false, nil
}

func lookupHost(host string, timeout int) ([]string, error) {
	c1 := make(chan []string, 1)
	e1 := make(chan error, 1)
	timeoutD := time.Duration(timeout) * time.Millisecond
	go func() {
		addrs, err := net.LookupHost(host)
		if err != nil {
			e1 <- err
		}
		c1 <- addrs
	}()
	select {
	case res := <-c1:
		return res, nil
	case err := <-e1:
		return nil, err
	case <-time.After(timeoutD):
		return nil, fmt.Errorf("DNS lookup timed out (%s)", timeoutD)
	}
}
