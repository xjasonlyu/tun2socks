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

func GetProcessName(addr net.Addr) string {
	// set default value
	var process = "N/A"
	if addr != nil {
		localHost, localPortStr, _ := net.SplitHostPort(addr.String())
		localPortInt, _ := strconv.Atoi(localPortStr)
		if cmd, _ := GetCommandNameBySocket(addr.Network(), localHost, uint16(localPortInt)); cmd != "" {
			process = cmd
		}
	}
	return process
}
