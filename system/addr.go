package system

import (
	"net"
	"strings"
	"time"
)

type Addr struct {
	address   string
	reachable bool
	Timeout   int64
}

func NewAddr(address string, system *System) *Addr {
	addr := normalizeAddress(address)
	return &Addr{address: addr}
}

func (h *Addr) Address() string {
	return h.address
}
func (h *Addr) Exists() (interface{}, error) { return h.Reachable() }

func (h *Addr) Reachable() (interface{}, error) {
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
