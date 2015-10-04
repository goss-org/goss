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
	sysPorts  map[string]GOnetstat.Process
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

func (p *Port) Exists() (interface{}, error) { return p.Listening() }

func (p *Port) Listening() (interface{}, error) {
	if _, ok := p.sysPorts[p.port]; ok {
		return true, nil
	}
	return false, nil
}

func (p *Port) IP() (interface{}, error) {
	return p.sysPorts[p.port].Ip, nil
}

func GetPorts(lookupPids bool) map[string]GOnetstat.Process {
	ports := make(map[string]GOnetstat.Process)
	netstat := GOnetstat.Tcp(lookupPids)
	var net string
	//netPorts := make(map[string]GOnetstat.Process)
	//ports["tcp"] = netPorts
	net = "tcp"
	for _, entry := range netstat {
		if entry.State == "LISTEN" {
			port := strconv.FormatInt(entry.Port, 10)
			ports[net+":"+port] = entry
		}
	}
	netstat = GOnetstat.Tcp6(lookupPids)
	//netPorts = make(map[string]GOnetstat.Process)
	//ports["tcp6"] = netPorts
	net = "tcp6"
	for _, entry := range netstat {
		if entry.State == "LISTEN" {
			port := strconv.FormatInt(entry.Port, 10)
			ports[net+":"+port] = entry
		}
	}
	netstat = GOnetstat.Udp(lookupPids)
	//netPorts = make(map[string]GOnetstat.Process)
	//ports["udp"] = netPorts
	net = "udp"
	for _, entry := range netstat {
		port := strconv.FormatInt(entry.Port, 10)
		ports[net+":"+port] = entry
	}
	netstat = GOnetstat.Udp6(lookupPids)
	//netPorts = make(map[string]GOnetstat.Process)
	//ports["udp6"] = netPorts
	net = "udp6"
	for _, entry := range netstat {
		port := strconv.FormatInt(entry.Port, 10)
		ports[net+":"+port] = entry
	}
	return ports
}
