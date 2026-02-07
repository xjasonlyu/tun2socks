package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func WithBindToInterface(iface *net.Interface) SocketOption {
	index := iface.Index
	return SocketOptionFunc(func(network, _ string, c syscall.RawConn) error {
		return rawConnControl(c, func(fd uintptr) error {
			switch network {
			case "ip4", "tcp4", "udp4":
				return unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF, index)
			case "ip6", "tcp6", "udp6":
				return unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_BOUND_IF, index)
			}
			return nil
		})
	})
}

func WithRoutingMark(_ int) SocketOption { return UnsupportedSocketOption }
