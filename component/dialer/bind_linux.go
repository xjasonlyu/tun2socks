package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func bindToInterface(i *net.Interface) controlFunc {
	return func(network, address string, c syscall.RawConn) error {
		ipStr, _, _ := net.SplitHostPort(address)
		if ip := net.ParseIP(ipStr); ip != nil && !ip.IsGlobalUnicast() {
			return nil
		}

		return c.Control(func(fd uintptr) {
			unix.BindToDevice(int(fd), i.Name)
		})
	}
}
