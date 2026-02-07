package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func WithInterface(iface *net.Interface) SocketOption {
	device := iface.Name
	return SocketOptionFunc(func(network, address string, c syscall.RawConn) error {
		return rawConnControl(c, func(fd uintptr) error {
			return unix.BindToDevice(int(fd), device)
		})
	})
}

func WithRoutingMark(mark int) SocketOption {
	return SocketOptionFunc(func(network, address string, c syscall.RawConn) error {
		return rawConnControl(c, func(fd uintptr) error {
			return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, mark)
		})
	})
}
