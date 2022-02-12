package tunnel

import (
	"net"
	"strconv"
)

// parseAddr parses net.Addr to IP and port.
func parseAddr(addr net.Addr) (net.IP, uint16) {
	switch v := addr.(type) {
	case *net.TCPAddr:
		return v.IP, uint16(v.Port)
	case *net.UDPAddr:
		return v.IP, uint16(v.Port)
	case nil:
		return nil, 0
	default:
		return parseAddrString(addr.String())
	}
}

// parseAddrString parses address string to IP and port.
func parseAddrString(addr string) (net.IP, uint16) {
	host, port, _ := net.SplitHostPort(addr)
	portInt, _ := strconv.ParseUint(port, 10, 16)
	return net.ParseIP(host), uint16(portInt)
}
