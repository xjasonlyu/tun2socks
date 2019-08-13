package lsof

import (
	"errors"
	"net"
	"strconv"

	"github.com/eycorsican/go-tun2socks/common/lsof"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrNotImplemented = errors.New("not implemented")
)

func GetProcessName(addr net.Addr) string {
	localHost, localPortStr, _ := net.SplitHostPort(addr.String())
	localPortInt, _ := strconv.Atoi(localPortStr)
	process, _ := lsof.GetCommandNameBySocket(addr.Network(), localHost, uint16(localPortInt))
	return process
}
