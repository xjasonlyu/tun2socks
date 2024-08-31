package tunnel

import (
	"net"
	"net/netip"

	"gvisor.dev/gvisor/pkg/tcpip"
)

// parseNetAddr parses net.Addr to IP and port.
func parseNetAddr(addr net.Addr) (netip.Addr, uint16) {
	if addr == nil {
		return netip.Addr{}, 0
	}
	if v, ok := addr.(interface {
		AddrPort() netip.AddrPort
	}); ok {
		ap := v.AddrPort()
		return ap.Addr(), ap.Port()
	}
	return parseAddrString(addr.String())
}

// parseAddrString parses address string to IP and port.
// It doesn't do any name resolution.
func parseAddrString(s string) (netip.Addr, uint16) {
	ap, err := netip.ParseAddrPort(s)
	if err != nil {
		return netip.Addr{}, 0
	}
	return ap.Addr(), ap.Port()
}

// parseTCPIPAddress parses tcpip.Address to netip.Addr.
func parseTCPIPAddress(addr tcpip.Address) netip.Addr {
	ip, _ := netip.AddrFromSlice(addr.AsSlice())
	return ip
}
