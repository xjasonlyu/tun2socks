package dialer

import (
	"net"
	"syscall"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/common/singledo"

	"golang.org/x/sys/unix"
)

var interfaces = singledo.NewSingle(30 * time.Second)

func resolveInterfaceByName(name string) (*net.Interface, error) {
	value, err, _ := interfaces.Do(func() (any, error) {
		return net.InterfaceByName(name)
	})
	if err != nil {
		return nil, err
	}
	return value.(*net.Interface), nil
}

func setSocketOptions(network, address string, c syscall.RawConn, opts *Options) (err error) {
	if opts == nil || !isTCPSocket(network) && !isUDPSocket(network) {
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
			var iface *net.Interface
			iface, innerErr = resolveInterfaceByName(opts.InterfaceName)
			if innerErr != nil {
				return
			}

			switch network {
			case "tcp4", "udp4":
				innerErr = unix.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_BOUND_IF, iface.Index)
			case "tcp6", "udp6":
				innerErr = unix.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_BOUND_IF, iface.Index)
			}
			if innerErr != nil {
				return
			}
		}
	})

	if innerErr != nil {
		err = innerErr
	}
	return
}
