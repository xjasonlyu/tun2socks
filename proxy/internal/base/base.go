package base

import (
	"context"
	"errors"
	"fmt"
	"net"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
)

var _ proxy.Proxy = (*Base)(nil)

type Base struct {
	address, protocol string
}

func New(address, protocol string) *Base {
	return &Base{
		address:  address,
		protocol: protocol,
	}
}

func (b *Base) Address() string {
	return b.address
}

func (b *Base) Protocol() string {
	return b.protocol
}

func (b *Base) String() string {
	return fmt.Sprintf("%s://%s", b.protocol, b.address)
}

func (b *Base) DialContext(context.Context, *M.Metadata) (net.Conn, error) {
	return nil, errors.ErrUnsupported
}

func (b *Base) DialUDP(*M.Metadata) (net.PacketConn, error) {
	return nil, errors.ErrUnsupported
}
