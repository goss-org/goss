package system

import (
	"fmt"
	"net"
	"regexp"
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
	Resolvable() (bool, error)
	Exists() (bool, error)
	Server() string
	Qtype() string
}

type DefDNS struct {
	host       string
	resolvable bool
	addrs      []string
	Timeout    int
	loaded     bool
	err        error
	server     string
	qtype      string
}

func NewDefDNS(host string, system *System, config util.Config) (DNS, error) {
	var h string
	var t string

	splitHost := strings.SplitN(host, ":", 2)
	if len(splitHost) == 2 && regexp.MustCompile(`^[A-Z]+$`).MatchString(splitHost[0]) {
		h = splitHost[1]
		t = splitHost[0]
	} else {
		h = host
	}

	return &DefDNS{
		host:    h,
		Timeout: config.TimeOutMilliSeconds(),
		server:  config.Server,
		qtype:   t,
	}, nil
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
			d.resolvable = false
			d.addrs = []string{}
			// DNSError is resolvable == false, ignore error
			if _, ok := err.(*net.DNSError); ok {
				return nil
			}
			d.err = err
			continue
		}
		sort.Strings(addrs)
		d.resolvable = true
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

func (d *DefDNS) Resolvable() (bool, error) {
	err := d.setup()

	return d.resolvable, err
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
			case "CAA":
				addrs, err = LookupCAA(host, server, c, m)
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

// A and AAAA record lookup - similar to net.LookupHost
func LookupHost(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	a, _ := LookupA(host, server, c, m)
	aaaa, _ := LookupAAAA(host, server, c, m)
	addrs = append(a, aaaa...)

	return
}

// A record lookup
func LookupA(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeA)
	r, _, err := c.Exchange(m, parseServerString(server))
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

// parseServerString - Check if the DNS Server in server config has a port, if not ensure 53 is prefixed.
func parseServerString(server string) string {
	srvhost, srvport, err := net.SplitHostPort(server)
	if err != nil {
		srvport = "53"
		srvhost = server
	}
	return net.JoinHostPort(srvhost, srvport)
}

// AAAA (IPv6) record lookup
func LookupAAAA(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeAAAA)
	r, _, err := c.Exchange(m, parseServerString(server))
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
	r, _, err := c.Exchange(m, parseServerString(server))
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
	r, _, err := c.Exchange(m, parseServerString(server))
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
	r, _, err := c.Exchange(m, parseServerString(server))
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
	r, _, err := c.Exchange(m, parseServerString(server))
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
	r, _, err := c.Exchange(m, parseServerString(server))
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

// PTR record lookup
func LookupPTR(addr string, server string, c *dns.Client, m *dns.Msg) (name []string, err error) {

	reverse, err := dns.ReverseAddr(addr)
	if err != nil {
		return nil, err
	}

	m.SetQuestion(reverse, dns.TypePTR)

	r, _, err := c.Exchange(m, parseServerString(server))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		name = append(name, ans.(*dns.PTR).Ptr)
	}

	return
}

// CAA record lookup
func LookupCAA(host string, server string, c *dns.Client, m *dns.Msg) (addrs []string, err error) {
	m.SetQuestion(dns.Fqdn(host), dns.TypeCAA)
	r, _, err := c.Exchange(m, parseServerString(server))
	if err != nil {
		return nil, err
	}

	for _, ans := range r.Answer {
		if t, ok := ans.(*dns.CAA); ok {
			flag := strconv.Itoa(int(t.Flag))
			caarec := strings.Join([]string{flag, t.Tag, t.Value}, " ")
			addrs = append(addrs, caarec)
		}
	}

	return
}
