package socks4

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/utils"
	"github.com/xjasonlyu/tun2socks/v2/transport/socks4"
)

var _ proxy.Proxy = (*Socks4)(nil)

type Socks4 struct {
	addr   string
	userID string
}

func New(addr, userID string) (*Socks4, error) {
	return &Socks4{
		addr:   addr,
		userID: userID,
	}, nil
}

func (ss *Socks4) DialContext(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	c, err = dialer.DialContext(ctx, "tcp", ss.addr)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", ss.addr, err)
	}
	utils.SetKeepAlive(c)

	defer func(c net.Conn) {
		utils.SafeConnClose(c, err)
	}(c)

	err = socks4.ClientHandshake(c, metadata.DestinationAddress(), socks4.CmdConnect, ss.userID)
	return c, err
}

func (ss *Socks4) DialUDP(*M.Metadata) (net.PacketConn, error) {
	return nil, errors.ErrUnsupported
}

func Parse(u *url.URL) (proxy.Proxy, error) {
	address, userID := u.Host, u.User.Username()
	return New(address, userID)
}

func init() {
	proxy.RegisterProtocol("socks4", Parse)
}
