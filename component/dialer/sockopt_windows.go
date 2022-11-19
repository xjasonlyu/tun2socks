package dialer

import (
	"encoding/binary"
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func setSocketOptions(network, address string, c syscall.RawConn, opts *Options) (err error) {
	if opts == nil || !isTCPSocket(network) && !isUDPSocket(network) {
		return
	}

	var innerErr error
	err = c.Control(func(fd uintptr) {
		host, _, _ := net.SplitHostPort(address)
		ip := net.ParseIP(host)
		if ip != nil && !ip.IsGlobalUnicast() {
			return
		}

		if opts.InterfaceName != "" {
			if len(ip) == net.IPv6len {
				innerErr = bindSocketToInterface6(windows.Handle(fd), uint32(opts.InterfaceIndex))
			} else {
				innerErr = bindSocketToInterface4(windows.Handle(fd), uint32(opts.InterfaceIndex))
			}
		}
	})

	if innerErr != nil {
		err = innerErr
	}
	return
}

func bindSocketToInterface4(handle windows.Handle, interfaceIndex uint32) error {
	const IP_UNICAST_IF = 31
	/* MSDN says for IPv4 this needs to be in net byte order, so that it's like an IP address with leading zeros. */
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], interfaceIndex)
	interfaceIndex = *(*uint32)(unsafe.Pointer(&bytes[0]))
	err := windows.SetsockoptInt(handle, windows.IPPROTO_IP, IP_UNICAST_IF, int(interfaceIndex))
	if err != nil {
		return err
	}
	return nil
}

func bindSocketToInterface6(handle windows.Handle, interfaceIndex uint32) error {
	const IPV6_UNICAST_IF = 31
	return windows.SetsockoptInt(handle, windows.IPPROTO_IPV6, IPV6_UNICAST_IF, int(interfaceIndex))
}
