package system

import (
	"strconv"
	"strings"

	"github.com/aelsabbahy/GOnetstat"
)

type Port struct {
	port      string
	listening bool
	ip        string
	sysPorts  map[string]map[string]string
}

func NewPort(port string, system *System) *Port {
	p := normalizePort(port)
	return &Port{
		port:     p,
		sysPorts: system.Ports(),
	}
}

func splitPort(fullport string) (network, port string) {
	split := strings.SplitN(fullport, ":", 2)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "tcp", fullport

}

func normalizePort(fullport string) string {
	net, addr := splitPort(fullport)
	return net + ":" + addr
}

func (p *Port) Port() string {
	return p.port
}

func (p *Port) Listening() (interface{}, error) {
	network, port := splitPort(p.port)
	if _, ok := p.sysPorts[network][port]; ok {
		return true, nil
	}
	return false, nil
}

func (p *Port) IP() (interface{}, error) {
	network, port := splitPort(p.port)
	return p.sysPorts[network][port], nil
}

func GetPorts() map[string]map[string]string {
	ports := make(map[string]map[string]string)
	netstat := GOnetstat.Tcp(false)
	netPorts := make(map[string]string)
	ports["tcp"] = netPorts
	for _, entry := range netstat {
		if entry.State == "LISTEN" {
			port := strconv.FormatInt(entry.Port, 10)
			netPorts[port] = entry.Ip
		}
	}
	netstat = GOnetstat.Tcp6(false)
	netPorts = make(map[string]string)
	ports["tcp6"] = netPorts
	for _, entry := range netstat {
		if entry.State == "LISTEN" {
			port := strconv.FormatInt(entry.Port, 10)
			netPorts[port] = entry.Ip
		}
	}
	netstat = GOnetstat.Udp(false)
	netPorts = make(map[string]string)
	ports["udp"] = netPorts
	for _, entry := range netstat {
		port := strconv.FormatInt(entry.Port, 10)
		netPorts[port] = entry.Ip
	}
	netstat = GOnetstat.Udp6(false)
	netPorts = make(map[string]string)
	ports["udp6"] = netPorts
	for _, entry := range netstat {
		port := strconv.FormatInt(entry.Port, 10)
		netPorts[port] = entry.Ip
	}
	return ports
}
