package engine

import (
	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
)

var _ adapter.Handler = (*fakeTunnel)(nil)

type fakeTunnel struct{}

func (*fakeTunnel) HandleTCP(conn adapter.TCPConn) {
	tunnel.TCPIn() <- conn
}

func (*fakeTunnel) HandleUDP(conn adapter.UDPConn) {
	tunnel.UDPIn() <- conn
}
