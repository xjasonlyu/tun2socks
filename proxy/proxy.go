// Package proxy provides implementations of proxy protocols.
package proxy

import (
	"context"
	"net"
	"time"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
)

const (
	DefaultConnectTimeout = 5 * time.Second
)

type Proxy interface {
	DialContext(context.Context, *M.Metadata) (net.Conn, error)
	DialUDP(*M.Metadata) (net.PacketConn, error)
}
