package socks5

import (
	"context"
	"net"

	"github.com/xjasonlyu/tun2socks/v2/dialer"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/proxy/internal/utils"
	"github.com/xjasonlyu/tun2socks/v2/transport/socks5"
)

// DialContextWithHandshake uses the default separate 5s connect / 21s handshake
// limits. The existing DialContext method retains its original semantics.
func (ss *Socks5) DialContextWithHandshake(ctx context.Context, metadata *M.Metadata) (c net.Conn, err error) {
	return ss.DialContextWithTimeouts(ctx, metadata, proxy.TCPTimeouts{})
}

// DialContextWithTimeouts bounds each stage separately. An earlier caller
// deadline or cancellation still takes precedence over both stage budgets.
func (ss *Socks5) DialContextWithTimeouts(ctx context.Context, metadata *M.Metadata, cfg proxy.TCPTimeouts) (net.Conn, error) {
	return ss.dialContextWithTimeouts(ctx, metadata, cfg, dialer.DialContext)
}

func (ss *Socks5) dialContextWithTimeouts(ctx context.Context, metadata *M.Metadata, cfg proxy.TCPTimeouts, dial func(context.Context, string, string) (net.Conn, error)) (c net.Conn, err error) {
	cfg, err = cfg.WithDefaults()
	if err != nil {
		return nil, err
	}
	network := "tcp"
	if ss.unix {
		network = "unix"
	}

	return utils.DialWithHandshake(ctx, network, ss.addr, "SOCKS5", cfg.ProxyHandshakeTimeout, dial, func(c net.Conn) error {
		var user *socks5.User
		if ss.user != "" {
			user = &socks5.User{Username: ss.user, Password: ss.pass}
		}
		_, err := socks5.ClientHandshake(c, utils.SerializeSocksAddr(metadata), socks5.CmdConnect, user)
		return err
	})
}
