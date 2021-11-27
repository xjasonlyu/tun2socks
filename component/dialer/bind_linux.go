package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func bindToInterface(i *net.Interface) controlFunc {
	return func(network, address string, c syscall.RawConn) (err error) {
		host, _, _ := net.SplitHostPort(address)
		if ip := net.ParseIP(host); ip != nil && !ip.IsGlobalUnicast() {
			return nil
		}

		var innerErr error
		err = c.Control(func(fd uintptr) {
			innerErr = unix.BindToDevice(int(fd), i.Name)
		})

		if innerErr != nil {
			err = innerErr
		}
		return
	}
}
