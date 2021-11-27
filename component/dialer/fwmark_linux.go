package dialer

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func setMark(m int) controlFunc {
	return func(_, _ string, c syscall.RawConn) (err error) {
		var innerErr error
		err = c.Control(func(fd uintptr) {
			innerErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, m)
		})

		if innerErr != nil {
			err = innerErr
		}
		return
	}
}
