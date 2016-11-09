package system

import (
	"fmt"
	"net"
	"sort"
	"time"
  "github.com/miekg/dns"

	"github.com/aelsabbahy/goss/util"
)

type DNS interface {
	Host() string
	Addrs() ([]string, error)
	Resolveable() (bool, error)
	Exists() (bool, error)
	Server() string
}

type DefDNS struct {
	host        string
	resolveable bool
	addrs       []string
	Timeout     int
	loaded      bool
	err         error
	server      string
}

func NewDefDNS(host string, system *System, config util.Config) DNS {
	return &DefDNS{
		host:    host,
		Timeout: config.Timeout,
		server:  config.Server,
	}
}

func (d *DefDNS) Host() string {
	return d.host
}

func (d *DefDNS) Server() string {
	return d.server
}

func (d *DefDNS) setup() error {
	if d.loaded {
		return d.err
	}
	d.loaded = true

	addrs, err := lookupHost(d.host, d.server, d.Timeout)
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

func (d *DefDNS) Resolveable() (bool, error) {
	err := d.setup()

	return d.resolveable, err
}

// Stub out
func (d *DefDNS) Exists() (bool, error) {
	return false, nil
}

func lookupHost(host string, server string, timeout int) ([]string, error) {

	c1 := make(chan []string, 1)
	e1 := make(chan error, 1)
	timeoutD := time.Duration(timeout) * time.Millisecond

	var addrs []string
	var err error
	go func() {
    if server != "" {
			addrs, err = serverLookup(host, server)
		} else {
		  addrs, err = net.LookupHost(host)
		}
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

func serverLookup(host string, server string) ([]string, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(host), dns.TypeA)

	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
    return nil, fmt.Errorf("%s", err)
  }
	if len(r.Answer) == 0 {
    return nil, fmt.Errorf("No DNS record found")
  }
	var answer []string
  for _, ans := range r.Answer {
    Arecord := ans.(*dns.A)
    answer = append(answer, Arecord.A.String())
  }
	return answer, nil
}
