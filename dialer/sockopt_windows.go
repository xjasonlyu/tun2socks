package dialer

import (
	"encoding/binary"
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	IP_UNICAST_IF   = 31
	IPV6_UNICAST_IF = 31
)

func WithInterface(iface *net.Interface) SocketOption {
	index := iface.Index
	return SocketOptionFunc(func(network, address string, c syscall.RawConn) error {
		return rawConnControl(c, func(fd uintptr) (err error) {
			switch network {
			case "ip4", "tcp4", "udp4":
				err = bindSocketToInterface4(windows.Handle(fd), uint32(index))
			case "ip6", "tcp6", "udp6":
				err = bindSocketToInterface6(windows.Handle(fd), uint32(index))
			}
			if network == "udp6" {
				// The underlying IP net maybe IPv4 even if the `network` param is `udp6`,
				// so we should bind socket to interface4 at the same time.
				_ = bindSocketToInterface4(windows.Handle(fd), uint32(index))
			}
			return
		})
	})
}

func bindSocketToInterface4(handle windows.Handle, index uint32) error {
	// For IPv4, this parameter must be an interface index in network byte order.
	// Ref: https://learn.microsoft.com/en-us/windows/win32/winsock/ipproto-ip-socket-options
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], index)
	index = *(*uint32)(unsafe.Pointer(&bytes[0]))
	return windows.SetsockoptInt(handle, windows.IPPROTO_IP, IP_UNICAST_IF, int(index))
}

func bindSocketToInterface6(handle windows.Handle, index uint32) error {
	return windows.SetsockoptInt(handle, windows.IPPROTO_IPV6, IPV6_UNICAST_IF, int(index))
}

func WithRoutingMark(_ int) SocketOption { return UnsupportedSocketOption }
