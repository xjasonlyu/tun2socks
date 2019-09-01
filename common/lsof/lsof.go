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

func isLocalIP(host string) bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	// all ip addresses
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			break
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && host == ip.String() {
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
		localHost, localPortStr, _ := net.SplitHostPort(addr.String())
		if isLocalIP(localHost) {
			localPortInt, _ := strconv.Atoi(localPortStr)
			if cmd, _ := GetCommandNameBySocket(addr.Network(), localHost, uint16(localPortInt)); cmd != "" {
				process = cmd
			}
		}
	}
	return process
}
