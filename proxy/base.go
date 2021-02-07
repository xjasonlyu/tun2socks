package proxy

import (
	"context"
	"errors"
	"net"

	"github.com/xjasonlyu/tun2socks/common/adapter"
)

var _ Proxy = (*Base)(nil)

type Base struct {
	addr string
}

func NewBase(addr string) *Base {
	return &Base{addr: addr}
}

func (b *Base) Addr() string {
	return b.addr
}

func (b *Base) Proto() string {
	return ""
}

func (b *Base) DialContext(context.Context, *adapter.Metadata) (net.Conn, error) {
	return nil, errors.New("not supported")
}

func (b *Base) DialUDP(*adapter.Metadata) (net.PacketConn, error) {
	return nil, errors.New("not supported")
}
