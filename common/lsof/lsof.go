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
	localHost, localPortStr, _ := net.SplitHostPort(addr.String())
	localPortInt, _ := strconv.Atoi(localPortStr)
	process, _ := GetCommandNameBySocket(addr.Network(), localHost, uint16(localPortInt))
	if process == "" {
		return "N/A"
	}
	return process
}
