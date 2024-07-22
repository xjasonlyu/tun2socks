package direct

import (
	"context"
	"net"
	"net/url"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal"
)

var _ proxy.Proxy = (*Direct)(nil)

const Protocol = "direct"

type Direct struct{ *internal.Base }

func New() *Direct { return &Direct{internal.New(Protocol, "")} }

func Parse(*url.URL) (proxy.Proxy, error) { return New(), nil }

func (d *Direct) DialContext(ctx context.Context, metadata *M.Metadata) (net.Conn, error) {
	c, err := dialer.DialContext(ctx, "tcp", metadata.DestinationAddress())
	if err != nil {
		return nil, err
	}
	internal.SetKeepAlive(c)
	return c, nil
}

func (d *Direct) DialUDP(*M.Metadata) (net.PacketConn, error) {
	pc, err := dialer.ListenPacket("udp", "")
	if err != nil {
		return nil, err
	}
	return &directPacketConn{PacketConn: pc}, nil
}

type directPacketConn struct {
	net.PacketConn
}

func (pc *directPacketConn) WriteTo(b []byte, addr net.Addr) (int, error) {
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		return pc.PacketConn.WriteTo(b, udpAddr)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr.String())
	if err != nil {
		return 0, err
	}
	return pc.PacketConn.WriteTo(b, udpAddr)
}

func init() {
	proxy.RegisterProtocol(Protocol, Parse)
}
