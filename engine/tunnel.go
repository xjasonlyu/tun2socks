package engine

import (
	"github.com/xjasonlyu/tun2socks/common/adapter"
	"github.com/xjasonlyu/tun2socks/component/stack"
	"github.com/xjasonlyu/tun2socks/tunnel"
)

var _ stack.Handler = (*fakeTunnel)(nil)

type fakeTunnel struct{}

func (*fakeTunnel) Add(conn adapter.TCPConn) {
	tunnel.Add(conn)
}

func (*fakeTunnel) AddPacket(packet adapter.UDPPacket) {
	tunnel.AddPacket(packet)
}
