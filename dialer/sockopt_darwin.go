package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func WithBindToInterface(iface *net.Interface) SocketOption {
	return SocketOptionFunc(func(network, address string, c syscall.RawConn) error {
		return control(c, func(fd uintptr) error {
			switch network {
			case "ip4", "tcp4", "udp4":
				return unix.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BOUND_IF, iface.Index)
			case "ip6", "tcp6", "udp6":
				return unix.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_BOUND_IF, iface.Index)
			}
			return nil
		})
	})
}

func WithRoutingMark(_ int) SocketOption { return NopSocketOption }
