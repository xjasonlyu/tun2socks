package proxy

import (
	"context"
	"errors"
	"net"

	"github.com/xjasonlyu/tun2socks/common/adapter"
	"github.com/xjasonlyu/tun2socks/proxy/proto"
)

var _ Proxy = (*Base)(nil)

type Base struct {
	addr  string
	proto proto.Proto
}

func NewBase(addr string, proto proto.Proto) *Base {
	return &Base{addr: addr, proto: proto}
}

func (b *Base) Addr() string {
	return b.addr
}

func (b *Base) Proto() proto.Proto {
	return b.proto
}

func (b *Base) DialContext(context.Context, *adapter.Metadata) (net.Conn, error) {
	return nil, errors.New("not supported")
}

func (b *Base) DialUDP(*adapter.Metadata) (net.PacketConn, error) {
	return nil, errors.New("not supported")
}
