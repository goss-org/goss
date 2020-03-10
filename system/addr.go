package system

import (
	"net"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/util"
)

type Addr interface {
	Address() string
	Exists() (bool, error)
	Reachable() (bool, error)
}

type DefAddr struct {
	address      string
	LocalAddress string
	Timeout      int
}

func NewDefAddr(address string, system *System, config util.Config) Addr {
	addr := normalizeAddress(address)
	return &DefAddr{
		address:      addr,
		LocalAddress: config.LocalAddress,
		Timeout:      config.Timeout,
	}
}

func (a *DefAddr) ID() string {
	return a.address
}
func (a *DefAddr) Address() string {
	return a.address
}
func (a *DefAddr) Exists() (bool, error) { return a.Reachable() }

func (a *DefAddr) Reachable() (bool, error) {
	network, address := splitAddress(a.address)

	var localAddr net.Addr
	if network == "udp" {
		localAddr = &net.UDPAddr{IP: net.ParseIP(a.LocalAddress)}
	} else {
		localAddr = &net.TCPAddr{IP: net.ParseIP(a.LocalAddress)}
	}
	d := net.Dialer{LocalAddr: localAddr, Timeout: time.Duration(a.Timeout) * time.Millisecond}
	conn, err := d.Dial(network, address)
	if err != nil {
		return false, nil
	}
	conn.Close()
	return true, nil
}

func splitAddress(fulladdress string) (network, address string) {
	split := strings.SplitN(fulladdress, "://", 2)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "tcp", fulladdress
}

func normalizeAddress(fulladdress string) string {
	net, addr := splitAddress(fulladdress)
	return net + "://" + addr
}
