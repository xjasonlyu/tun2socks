package lsof

import (
	"errors"
	"net"
	"strconv"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotImplemented = errors.New("not implemented")
)

func isLocalIP(ip net.IP) bool {
	interfaces, _ := net.Interfaces()
	// all ip addresses
	for _, i := range interfaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ifIP net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ifIP = v.IP
			case *net.IPAddr:
				ifIP = v.IP
			}
			if ifIP != nil && ip.Equal(ifIP) {
				return true
			}
		}
	}
	return false
}

func GetProcessName(addr net.Addr) string {
	// set default value
	var process = "N/A"
	if addr != nil {
		localIP, localPort, _ := net.SplitHostPort(addr.String())
		ip := net.ParseIP(localIP)
		// ignore non-local IP
		if ip.IsUnspecified() || ip.IsLoopback() || isLocalIP(ip) {
			port, _ := strconv.Atoi(localPort)
			if cmd, _ := GetCommandNameBySocket(addr.Network(), localIP, uint16(port)); cmd != "" {
				process = cmd
			}
		}
	}
	return process
}
