package engine

import (
	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
)

var _ adapter.Handler = (*fakeTunnel)(nil)

type fakeTunnel struct{}

func (*fakeTunnel) HandleTCPConn(conn adapter.TCPConn) {
	tunnel.TCPIn() <- conn
}

func (*fakeTunnel) HandleUDPConn(conn adapter.UDPConn) {
	tunnel.UDPIn() <- conn
}
