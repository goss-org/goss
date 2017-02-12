package system

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/util"
	"github.com/miekg/dns"
)

type DNS interface {
	Host() string
	Addrs() ([]string, error)
	Resolveable() (bool, error)
	Exists() (bool, error)
	Server() string
	Qtype() string
}

type DefDNS struct {
	host        string
	resolveable bool
	addrs       []string
	Timeout     int
	loaded      bool
	err         error
	server      string
	qtype       string
}

func NewDefDNS(host string, system *System, config util.Config) DNS {
	var h string
	var t string
	if len(strings.Split(host, ":")) > 1 {
		h = strings.Split(host, ":")[1]
		t = strings.Split(host, ":")[0]
	} else {
		h = host
	}
	return &DefDNS{
		host:    h,
		Timeout: config.Timeout,
		server:  config.Server,
		qtype:   t,
	}
}

func (d *DefDNS) Host() string {
	return d.host
}

func (d *DefDNS) Server() string {
	return d.server
}

func (d *DefDNS) Qtype() string {
	return d.qtype
}

func (d *DefDNS) setup() error {
	if d.loaded {
		return d.err
	}
	d.loaded = true

	for i := 0; i < 3; i++ {
		addrs, err := DNSlookup(d.host, d.server, d.qtype, d.Timeout)
		if err != nil || len(addrs) == 0 {
			d.resolveable = false
			d.addrs = []string{}
			// DNSError is resolvable == false, ignore error
			if _, ok := err.(*net.DNSError); ok {
				return nil
			}
			d.err = err
			continue
		}
		sort.Strings(addrs)
		d.resolveable = true
		d.addrs = addrs
		d.err = nil
		return nil
	}
	return d.err
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

func DNSlookup(host string, server string, qtype string, timeout int) ([]string, error) {
	c1 := make(chan []string, 1)
	e1 := make(chan error, 1)
	timeoutD := time.Duration(timeout) * time.Millisecond

	var addrs []string
	var err error
	go func() {
		if server != "" {
			c := new(dns.Client)
			c.Timeout = timeoutD
			m := new(dns.Msg)

			switch qtype {
			case "A":
				addrs, err = LookupA(host, server, c, m)
			case "AAAA":
				addrs, err = LookupAAAA(host, server, c, m)
			case "PTR":
				addrs, err = LookupPTR(host, server, c, m)
			case "CNAME":
				addrs, err = LookupCNAME(host, server, c, m)
			case "MX":
				addrs, err = LookupMX(host, server, c, m)
			case "NS":
				addrs, err = LookupNS(host, server, c, m)
			case "SRV":
				addrs, err = LookupSRV(host, server, c, m)
			case "TXT":
				addrs, err = LookupTXT(host, server, c, m)
			default:
				addrs, err = LookupHost(host, server, c, m)
			}
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

// These are function are a re-implementation of the net.Lookup* ones
// They are adapted to the package miekg/dns.

// LookupAddr performs a reverse lookup for the given address, returning a
// list of names mapping to that address.
func LookupPTR(addr string, server string, c *dns.Client, m *dns.Msg) (name []string, err error) {

	reverse, err := dns.ReverseAddr(addr)
	if err != nil {
		return nil, err
	}

	m.SetQuestion(reverse, dns.TypePTR)

	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return nil, err
	}
	for _, ans := range r.Answer {
		name = append(name, ans.(*dns.PTR).Ptr)
	}
	return
}

// LookupHost looks up the given host. It returns
// an array of that host's addresses IPv4 and IPv6.
func LookupHost(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	a, _ := LookupA(host, server, c, m)
	aaaa, _ := LookupAAAA(host, server, c, m)
	addrs = append(a, aaaa...)

	return
}

// A record lookup
func LookupA(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeA)
	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.A); ok {
			addrs = append(addrs, t.A.String())
		}
	}

	return
}

// AAAA (IPv6) record lookup
func LookupAAAA(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeAAAA)
	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.AAAA); ok {
			addrs = append(addrs, t.AAAA.String())
		}
	}

	return
}

// CNAME record lookup
func LookupCNAME(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeCNAME)
	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.CNAME); ok {
			addrs = append(addrs, t.Target)
		}
	}

	return
}

// MX record lookup
func LookupMX(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeMX)
	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.MX); ok {
			mxstring := strconv.Itoa(int(t.Preference)) + " " + t.Mx
			addrs = append(addrs, mxstring)
		}
	}

	return
}

// NS record lookup
func LookupNS(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeNS)
	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.NS); ok {
			addrs = append(addrs, t.Ns)
		}
	}

	return
}

// SRV record lookup
func LookupSRV(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeSRV)
	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.SRV); ok {
			prio := strconv.Itoa(int(t.Priority))
			weight := strconv.Itoa(int(t.Weight))
			port := strconv.Itoa(int(t.Port))
			srvrec := strings.Join([]string{prio, weight, port, t.Target}, " ")
			addrs = append(addrs, srvrec)
		}
	}

	return
}

// TXT record lookup
func LookupTXT(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeTXT)
	r, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.TXT); ok {
			addrs = append(addrs, t.Txt...)
		}
	}

	return
}
