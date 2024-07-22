package socks4

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/base"
	"github.com/xjasonlyu/tun2socks/v2/transport/socks4"
)

var _ proxy.Proxy = (*Socks4)(nil)

const protocol = "socks4"

type Socks4 struct {
	*base.Base

	userID string
}

func New(addr, userID string) (*Socks4, error) {
	return &Socks4{
		Base:   base.New(addr, protocol),
		userID: userID,
	}, nil
}

func Parse(proxyURL *url.URL) (proxy.Proxy, error) {
	address, userID := proxyURL.Host, proxyURL.User.Username()
	return New(address, userID)
}

func (ss *Socks4) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	c, err = dialer.DialContext(ctx, "tcp", ss.Address())
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", ss.Address(), err)
	}
	internal.SetKeepAlive(c)

	defer func(c net.Conn) {
		internal.SafeConnClose(c, err)
	}(c)

	err = socks4.ClientHandshake(c, metadata.DestinationAddress(), socks4.CmdConnect, ss.userID)
	return
}

func init() {
	proxy.RegisterProtocol(protocol, Parse)
}
