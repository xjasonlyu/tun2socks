package engine

import (
	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
)

var _ core.Handler = (*fakeTunnel)(nil)

type fakeTunnel struct{}

func (*fakeTunnel) Add(conn core.TCPConn) {
	tunnel.Add(conn)
}

func (*fakeTunnel) AddPacket(packet core.UDPPacket) {
	tunnel.AddPacket(packet)
}
