package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func WithBindToInterface(_ *net.Interface) SocketOption { return UnsupportedSocketOption }

func WithRoutingMark(mark int) SocketOption {
	return SocketOptionFunc(func(_, _ string, c syscall.RawConn) error {
		return rawConnControl(c, func(fd uintptr) error {
			return unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_USER_COOKIE, mark)
		})
	})
}
