package tunnel

import (
	"net"

	"github.com/xjasonlyu/tun2socks/common/adapter"
	"github.com/xjasonlyu/tun2socks/component/manager"
)

var (
	// DefaultManager is the default traffic and connections
	// manager used by tunnel.
	DefaultManager = manager.New()
)

func newTCPTracker(conn net.Conn, metadata *adapter.Metadata) net.Conn {
	return manager.NewTCPTracker(conn, metadata, DefaultManager)
}

func newUDPTracker(conn net.PacketConn, metadata *adapter.Metadata) net.PacketConn {
	return manager.NewUDPTracker(conn, metadata, DefaultManager)
}
