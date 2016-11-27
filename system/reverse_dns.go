package system

import (
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/aelsabbahy/goss/util"
)

type ReverseDNS interface {
	Addr() string
	Hosts() ([]string, error)
	Resolveable() (bool, error)
	Exists() (bool, error)
}

type DefReverseDNS struct {
	addr        string
	resolveable bool
	hosts       []string
	Timeout     int
	loaded      bool
	err         error
}

func NewDefReverseDNS(addr string, system *System, config util.Config) ReverseDNS {
	return &DefReverseDNS{
		addr:    addr,
		Timeout: config.Timeout,
	}
}

func (d *DefReverseDNS) Addr() string {
	return d.addr
}

func (d *DefReverseDNS) setup() error {
	if d.loaded {
		return d.err
	}
	d.loaded = true

	hosts, err := lookupAddr(d.addr, d.Timeout)
	if err != nil || len(hosts) == 0 {
		d.resolveable = false
		d.hosts = []string{}
		// DNSError is resolvable == false, ignore error
		if _, ok := err.(*net.DNSError); ok {
			return nil
		}
		d.err = err
		return d.err
	}
	sort.Strings(hosts)
	d.resolveable = true
	d.hosts = hosts
	return nil
}

func (d *DefReverseDNS) Hosts() ([]string, error) {
	err := d.setup()

	return d.hosts, err
}

func (d *DefReverseDNS) Resolveable() (bool, error) {
	err := d.setup()

	return d.resolveable, err
}

// Stub out
func (d *DefReverseDNS) Exists() (bool, error) {
	return false, nil
}

func lookupAddr(addr string, timeout int) ([]string, error) {
	c1 := make(chan []string, 1)
	e1 := make(chan error, 1)
	timeoutD := time.Duration(timeout) * time.Millisecond
	go func() {
		hosts, err := net.LookupAddr(addr)
		if err != nil {
			e1 <- err
		}
		c1 <- hosts
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
