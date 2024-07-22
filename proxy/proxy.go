// Package proxy provides implementations of proxy protocols.
package proxy

import (
	"context"
	"net"
	"time"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
)

const (
	TCPConnectTimeout = 5 * time.Second
)

// DefaultProxy is the default [Proxy] and is used by [Dial], [DialContext], and [DialUDP].
var DefaultProxy Proxy = nil

type Proxy interface {
	// Address returns the address of the proxy.
	Address() string

	// Protocol returns the protocol of the proxy.
	Protocol() string

	// String returns the string representation of the proxy.
	String() string

	// DialContext is used to dial TCP networks with context.
	DialContext(context.Context, *M.Metadata) (net.Conn, error)

	// DialUDP is used to to dial/listen UDP networks.
	DialUDP(*M.Metadata) (net.PacketConn, error)
}

// Dial uses the DefaultProxy to dial TCP.
func Dial(metadata *M.Metadata) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), TCPConnectTimeout)
	defer cancel()
	return DialContext(ctx, metadata)
}

// DialContext uses the DefaultProxy to dial TCP with context.
func DialContext(ctx context.Context, metadata *M.Metadata) (net.Conn, error) {
	return DefaultProxy.DialContext(ctx, metadata)
}

// DialUDP uses the DefaultProxy to dial UDP.
func DialUDP(metadata *M.Metadata) (net.PacketConn, error) {
	return DefaultProxy.DialUDP(metadata)
}
