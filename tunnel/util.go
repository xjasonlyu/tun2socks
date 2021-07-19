package tunnel

import (
	"net"
	"strconv"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// parseAddr parses address to IP and port.
func parseAddr(addr string) (net.IP, uint16) {
	host, portStr, _ := net.SplitHostPort(addr)
	portInt, _ := strconv.ParseUint(portStr, 10, 16)
	return net.ParseIP(host), uint16(portInt)
}
