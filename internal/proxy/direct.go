package proxy

import (
	"context"
	"net"
	"net/url"

	"github.com/xjasonlyu/clash/component/dialer"
	"github.com/xjasonlyu/tun2socks/internal/adapter"
)

type Direct struct {
	*Base
}

func NewDirect(url *url.URL) (*Direct, error) {
	return &Direct{
		Base: &Base{
			url: url,
		},
	}, nil
}

func (d *Direct) DialContext(ctx context.Context, metadata *adapter.Metadata) (net.Conn, error) {
	c, err := dialer.DialContext(ctx, "tcp", metadata.DestinationAddress())
	if err != nil {
		return nil, err
	}
	tcpKeepAlive(c)
	return c, nil
}

func (d *Direct) DialUDP(_ *adapter.Metadata) (net.PacketConn, error) {
	pc, err := dialer.ListenPacket("udp", "")
	if err != nil {
		return nil, err
	}
	return &directPacketConn{PacketConn: pc}, nil
}

type directPacketConn struct {
	net.PacketConn
}

func (pc *directPacketConn) WriteTo(b []byte, addr net.Addr) (_ int, err error) {
	var udpAddr *net.UDPAddr
	if m, ok := addr.(*adapter.Metadata); ok && m.Host == "" {
		udpAddr = m.UDPAddr()
	} else {
		udpAddr, err = resolveUDPAddr("udp", addr.String())
	}

	if err != nil {
		return 0, err
	}
	return pc.PacketConn.WriteTo(b, udpAddr)
}
