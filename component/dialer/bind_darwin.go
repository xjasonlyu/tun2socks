package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func bindToInterface(network, address string, c syscall.RawConn) error {
	ipStr, _, _ := net.SplitHostPort(address)
	if ip := net.ParseIP(ipStr); ip != nil && !ip.IsGlobalUnicast() {
		return nil
	}

	return c.Control(func(fd uintptr) {
		switch network {
		case "tcp4", "udp4":
			unix.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BOUND_IF, _boundInterface.Index)
		case "tcp6", "udp6":
			unix.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_BOUND_IF, _boundInterface.Index)
		}
	})
}
