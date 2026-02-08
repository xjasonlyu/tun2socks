package dialer

import (
	"math/bits"
	"net"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	IP_UNICAST_IF   = 0x1f
	IPV6_UNICAST_IF = 0x1f
)

func WithBindToInterface(iface *net.Interface) SocketOption {
	index := uint32(iface.Index)
	return SocketOptionFunc(func(network, address string, c syscall.RawConn) error {
		return rawConnControl(c, func(fd uintptr) (err error) {
			switch network {
			case "ip4", "tcp4", "udp4":
				err = bindSocketToInterface4(windows.Handle(fd), index)
			case "ip6", "tcp6", "udp6":
				err = bindSocketToInterface6(windows.Handle(fd), index)
				// UDPv6 may still use an IPv4 underlying socket if the destination
				// address is unspecified (e.g. ":0").
				if network == "udp6" {
					host, _, _ := net.SplitHostPort(address)
					if ip := net.ParseIP(host); ip == nil || ip.IsUnspecified() {
						_ = bindSocketToInterface4(windows.Handle(fd), index)
					}
				}
			}
			return
		})
	})
}

func bindSocketToInterface4(handle windows.Handle, index uint32) error {
	// For IPv4, this parameter must be an interface index in network byte order.
	// Ref: https://learn.microsoft.com/en-us/windows/win32/winsock/ipproto-ip-socket-options
	index = bits.ReverseBytes32(index)
	return windows.SetsockoptInt(handle, windows.IPPROTO_IP, IP_UNICAST_IF, int(index))
}

func bindSocketToInterface6(handle windows.Handle, index uint32) error {
	return windows.SetsockoptInt(handle, windows.IPPROTO_IPV6, IPV6_UNICAST_IF, int(index))
}

func WithRoutingMark(_ int) SocketOption { return UnsupportedSocketOption }
