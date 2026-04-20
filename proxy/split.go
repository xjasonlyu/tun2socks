package proxy

import (
	"context"
	"errors"
	"net"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
)

// Split pairs two independent proxies so TCP and UDP flows can be routed
// through different backends. This is required for UDP-only backends such
// as MASQUE (RFC 9298) alongside any TCP-capable proxy.
type Split struct {
	TCP Proxy
	UDP Proxy
}

var _ Proxy = (*Split)(nil)

func (s *Split) DialContext(ctx context.Context, metadata *M.Metadata) (net.Conn, error) {
	if s.TCP == nil {
		return nil, errors.New("split: no TCP proxy configured")
	}
	return s.TCP.DialContext(ctx, metadata)
}

func (s *Split) DialUDP(metadata *M.Metadata) (net.PacketConn, error) {
	if s.UDP == nil {
		return nil, errors.New("split: no UDP proxy configured")
	}
	return s.UDP.DialUDP(metadata)
}
