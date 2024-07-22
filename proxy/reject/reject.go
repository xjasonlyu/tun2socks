package reject

import (
	"context"
	"io"
	"net"
	"net/url"
	"time"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/base"
)

var _ proxy.Proxy = (*Reject)(nil)

const protocol = "reject"

type Reject struct{ *base.Base }

func New() *Reject { return &Reject{base.New("", protocol)} }

func Parse(*url.URL) (proxy.Proxy, error) { return New(), nil }

func (r *Reject) DialContext(context.Context, *M.Metadata) (net.Conn, error) { return &nopConn{}, nil }

func (r *Reject) DialUDP(*M.Metadata) (net.PacketConn, error) { return &nopPacketConn{}, nil }

type nopConn struct{}

func (rw *nopConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (rw *nopConn) Write([]byte) (int, error)        { return 0, io.EOF }
func (rw *nopConn) Close() error                     { return nil }
func (rw *nopConn) LocalAddr() net.Addr              { return nil }
func (rw *nopConn) RemoteAddr() net.Addr             { return nil }
func (rw *nopConn) SetDeadline(time.Time) error      { return nil }
func (rw *nopConn) SetReadDeadline(time.Time) error  { return nil }
func (rw *nopConn) SetWriteDeadline(time.Time) error { return nil }

type nopPacketConn struct{}

func (npc *nopPacketConn) WriteTo(b []byte, _ net.Addr) (n int, err error) { return len(b), nil }
func (npc *nopPacketConn) ReadFrom([]byte) (int, net.Addr, error)          { return 0, nil, io.EOF }
func (npc *nopPacketConn) Close() error                                    { return nil }
func (npc *nopPacketConn) LocalAddr() net.Addr                             { return &net.UDPAddr{IP: net.IPv4zero, Port: 0} }
func (npc *nopPacketConn) SetDeadline(time.Time) error                     { return nil }
func (npc *nopPacketConn) SetReadDeadline(time.Time) error                 { return nil }
func (npc *nopPacketConn) SetWriteDeadline(time.Time) error                { return nil }

func init() {
	proxy.RegisterProtocol(protocol, Parse)
}
