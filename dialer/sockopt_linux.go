package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func WithBindToInterface(iface *net.Interface) SocketOption {
	return SocketOptionFunc(func(network, address string, c syscall.RawConn) error {
		return control(c, func(fd uintptr) error {
			return unix.BindToDevice(int(fd), iface.Name)
		})
	})
}

func WithRoutingMark(mark int) SocketOption {
	return SocketOptionFunc(func(network, address string, c syscall.RawConn) error {
		return control(c, func(fd uintptr) error {
			return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, mark)
		})
	})
}
