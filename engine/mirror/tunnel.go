package mirror

import (
	"github.com/TianHe-Labs/Zeus/core/adapter"
	"github.com/TianHe-Labs/Zeus/tunnel"
)

var _ adapter.TransportHandler = (*Tunnel)(nil)

type Tunnel struct{}

func (*Tunnel) HandleTCP(conn adapter.TCPConn) {
	tunnel.TCPIn() <- conn
}

func (*Tunnel) HandleUDP(conn adapter.UDPConn) {
	tunnel.UDPIn() <- conn
}
