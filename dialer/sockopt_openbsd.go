package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func WithBindToInterface(_ *net.Interface) SocketOption { return NopSocketOption }

func WithRoutingMark(mark int) SocketOption {
	return SocketOptionFunc(func(network, address string, c syscall.RawConn) error {
		return control(c, func(fd uintptr) error {
			return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_RTABLE, mark)
		})
	})
}
