package system

import (
	"net"
	"strings"
	"time"
)

type Addr interface {
	Address() string
	Exists() (interface{}, error)
	Reachable() (interface{}, error)
	SetTimeout(int64)
}

type DefAddr struct {
	address   string
	reachable bool
	Timeout   int64
}

func NewDefAddr(address string, system *System) Addr {
	addr := normalizeAddress(address)
	return &DefAddr{address: addr}
}

func (h *DefAddr) SetTimeout(t int64) {
	h.Timeout = t
}

func (h *DefAddr) ID() string {
	return h.address
}
func (h *DefAddr) Address() string {
	return h.address
}
func (h *DefAddr) Exists() (interface{}, error) { return h.Reachable() }

func (h *DefAddr) Reachable() (interface{}, error) {
	network, address := splitAddress(h.address)
	timeout := h.Timeout
	if timeout == 0 {
		timeout = 500
	}
	conn, err := net.DialTimeout(network, address, time.Duration(timeout)*time.Millisecond)
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
