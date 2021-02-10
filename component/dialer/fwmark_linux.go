package dialer

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func setMark(i int) controlFunc {
	return func(_, _ string, c syscall.RawConn) error {
		return c.Control(func(fd uintptr) {
			unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, i)
		})
	}
}
