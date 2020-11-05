package proxy

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/xjasonlyu/tun2socks/internal/adapter"
)

type Base struct {
	url *url.URL
}

func NewBase(url *url.URL) (*Base, error) {
	return &Base{
		url: url,
	}, nil
}

func (b *Base) Type() string {
	if b.url == nil {
		return ""
	}
	return strings.ToLower(b.url.Scheme)
}

func (b *Base) Addr() string {
	if b.url == nil {
		return ""
	}
	return b.url.Host
}

func (b *Base) String() string {
	if b.url == nil {
		return ""
	}
	return b.url.String()
}

func (b *Base) DialContext(context.Context, *adapter.Metadata) (net.Conn, error) {
	return nil, errors.New("no support")
}

func (b *Base) DialUDP(*adapter.Metadata) (net.PacketConn, error) {
	return nil, errors.New("no support")
}
