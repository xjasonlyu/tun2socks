// Package proxy provides implementations of proxy protocols.
package proxy

import (
	"context"
	"net"
	"time"

	"github.com/xjasonlyu/tun2socks/common/adapter"
)

const (
	tcpConnectTimeout = 5 * time.Second
)

var (
	_defaultDialer Dialer = &Base{}
)

type Dialer interface {
	DialContext(context.Context, *adapter.Metadata) (net.Conn, error)
	DialUDP(*adapter.Metadata) (net.PacketConn, error)
}

type Proxy interface {
	Dialer
	Addr() string
	Proto() string
}

// SetDialer sets default Dialer.
func SetDialer(d Dialer) {
	_defaultDialer = d
}

// Dial uses default Dialer to dial TCP.
func Dial(metadata *adapter.Metadata) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), tcpConnectTimeout)
	defer cancel()
	return _defaultDialer.DialContext(ctx, metadata)
}

// DialContext uses default Dialer to dial TCP with context.
func DialContext(ctx context.Context, metadata *adapter.Metadata) (net.Conn, error) {
	return _defaultDialer.DialContext(ctx, metadata)
}

// DialUDP uses default Dialer to dial UDP.
func DialUDP(metadata *adapter.Metadata) (net.PacketConn, error) {
	return _defaultDialer.DialUDP(metadata)
}
