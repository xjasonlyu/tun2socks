package proxy

import (
	"context"
	"fmt"
	"net"

	"github.com/TianHe-Labs/Zeus/dialer"
	M "github.com/TianHe-Labs/Zeus/metadata"
	"github.com/TianHe-Labs/Zeus/proxy/proto"
	"github.com/TianHe-Labs/Zeus/transport/socks4"
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

	defer safeConnClose(c, err)

	err = socks4.ClientHandshake(c, metadata.DestinationAddress(), socks4.CmdConnect, ss.userID)
	return
}
