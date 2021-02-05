package stack

import (
	"github.com/xjasonlyu/tun2socks/common/adapter"
)

type Handler interface {
	Add(adapter.TCPConn)
	AddPacket(adapter.UDPPacket)
}
