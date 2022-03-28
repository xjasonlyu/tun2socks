package dialer

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func setSocketOptions(network, address string, c syscall.RawConn, opts *Options) (err error) {
	if !isTCPSocket(network) && !isUDPSocket(network) {
		return
	}

	var innerErr error
	err = c.Control(func(fd uintptr) {
		// must be GlobalUnicast.
		host, _, _ := net.SplitHostPort(address)
		if ip := net.ParseIP(host); ip != nil && !ip.IsGlobalUnicast() {
			return
		}

		if opts.InterfaceName != "" {
			if innerErr = unix.BindToDevice(int(fd), opts.InterfaceName); innerErr != nil {
				return
			}
		}
		if opts.RoutingMark != 0 {
			if innerErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, opts.RoutingMark); innerErr != nil {
				return
			}
		}
	})

	if innerErr != nil {
		err = innerErr
	}
	return
}
