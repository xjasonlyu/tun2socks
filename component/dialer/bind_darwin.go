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
			return
		}

		var innerErr error
		err = c.Control(func(fd uintptr) {
			switch network {
			case "tcp4", "udp4":
				innerErr = unix.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BOUND_IF, i.Index)
			case "tcp6", "udp6":
				innerErr = unix.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_BOUND_IF, i.Index)
			}
		})

		if innerErr != nil {
			err = innerErr
		}
		return
	}
}
