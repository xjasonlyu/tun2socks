package engine

import (
	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
)

var _ core.Handler = (*fakeTunnel)(nil)

type fakeTunnel struct{}

func (*fakeTunnel) HandleTCPConn(conn core.TCPConn) {
	tunnel.TCPIn() <- conn
}

func (*fakeTunnel) HandleUDPConn(conn core.UDPConn) {
	tunnel.UDPIn() <- conn
}
