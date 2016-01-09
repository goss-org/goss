package system

import (
	"net"
	"strings"
	"time"

	"github.com/aelsabbahy/goss/util"
)

type Addr interface {
	Address() string
	Exists() (interface{}, error)
	Reachable() (interface{}, error)
}

type DefAddr struct {
	address string
	Timeout int
}

func NewDefAddr(address string, system *System, config util.Config) Addr {
	addr := normalizeAddress(address)
	return &DefAddr{
		address: addr,
		Timeout: config.Timeout,
	}
}

func (a *DefAddr) ID() string {
	return a.address
}
func (a *DefAddr) Address() string {
	return a.address
}
func (a *DefAddr) Exists() (interface{}, error) { return a.Reachable() }

func (a *DefAddr) Reachable() (interface{}, error) {
	network, address := splitAddress(a.address)

	conn, err := net.DialTimeout(network, address, time.Duration(a.Timeout)*time.Millisecond)
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
