package proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy/proto"
	"github.com/xjasonlyu/tun2socks/v2/transport/socks4"
)

var _ Proxy = (*Socks4)(nil)

type Socks4 struct {
	*Base

	userID string
}

func NewSocks4(addr, userID string) (*Socks4, error) {
	return &Socks4{
		Base: &Base{
			addr:  addr,
			proto: proto.Socks4,
		},
		userID: userID,
	}, nil
}

func (ss *Socks4) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	c, err = dialer.DialContext(ctx, "tcp", ss.Addr())
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", ss.Addr(), err)
	}
	setKeepAlive(c)

	defer func(c net.Conn) {
		safeConnClose(c, err)
	}(c)

	err = socks4.ClientHandshake(c, metadata.DestinationAddress(), socks4.CmdConnect, ss.userID)
	return c, err
}
