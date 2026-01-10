// Package proxy provides implementations of proxy protocols.
package proxy

import (
	"context"
	"net"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
)

type Proxy interface {
	DialContext(context.Context, *M.Metadata) (net.Conn, error)
	DialUDP(*M.Metadata) (net.PacketConn, error)
}
