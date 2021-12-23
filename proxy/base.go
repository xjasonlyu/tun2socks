package proxy

import (
	"context"
	"errors"
	"net"

	M "github.com/xjasonlyu/tun2socks/v2/constant"
	"github.com/xjasonlyu/tun2socks/v2/proxy/proto"
)

var _ Proxy = (*Base)(nil)

type Base struct {
	addr  string
	proto proto.Proto
}

func (b *Base) Addr() string {
	return b.addr
}

func (b *Base) Proto() proto.Proto {
	return b.proto
}

func (b *Base) DialContext(context.Context, *M.Metadata) (net.Conn, error) {
	return nil, errors.New("not supported")
}

func (b *Base) DialUDP(*M.Metadata) (net.PacketConn, error) {
	return nil, errors.New("not supported")
}
