package core

import (
	"fmt"
	"net"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"

	"github.com/xjasonlyu/tun2socks/internal/adapter"
	"github.com/xjasonlyu/tun2socks/pkg/log"
)

const (
	// defaultWndSize if set to zero, the default
	// receive window buffer size is used instead.
	defaultWndSize = 0

	// maxConnAttempts specifies the maximum number
	// of in-flight tcp connection attempts.
	maxConnAttempts = 2 << 10

	// tcpKeepaliveIdle specifies the time a connection
	// must remain idle before the first TCP keepalive
	// packet is sent. Once this time is reached,
	// tcpKeepaliveInterval option is used instead.
	tcpKeepaliveIdle = 60 * time.Second

	// tcpKeepaliveInterval specifies the interval
	// time between sending TCP keepalive packets.
	tcpKeepaliveInterval = 30 * time.Second
)

type tcpHandleFunc func(adapter.TCPConn)

func WithTCPHandler(handle tcpHandleFunc) Option {
	return func(s *stack.Stack) error {
		tcpForwarder := tcp.NewForwarder(s, defaultWndSize, maxConnAttempts, func(r *tcp.ForwarderRequest) {
			var wq waiter.Queue
			id := r.ID()
			ep, err := r.CreateEndpoint(&wq)
			if err != nil {
				log.Warnf("[STACK] %s create endpoint error: %v", formatID(&id), err)
				// prevent potential half-open TCP connection leak.
				r.Complete(true)
				return
			}
			r.Complete(false)

			if err := setKeepalive(ep); err != nil {
				log.Warnf("[STACK] %s %v", formatID(&id), err)
			}

			conn := &tcpConn{
				Conn: gonet.NewTCPConn(&wq, ep),
				metadata: &adapter.Metadata{
					Net:     adapter.TCP,
					SrcIP:   net.IP(id.RemoteAddress),
					SrcPort: id.RemotePort,
					DstIP:   net.IP(id.LocalAddress),
					DstPort: id.LocalPort,
				},
			}

			handle(conn)
		})
		s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)
		return nil
	}
}

func formatID(id *stack.TransportEndpointID) string {
	return fmt.Sprintf(
		"%s:%d --> %s:%d",
		id.RemoteAddress,
		id.RemotePort,
		id.LocalAddress,
		id.LocalPort,
	)
}

func setKeepalive(ep tcpip.Endpoint) error {
	if err := ep.SetSockOptBool(tcpip.KeepaliveEnabledOption, true); err != nil {
		return fmt.Errorf("set keepalive: %s", err)
	}
	idleOpt := tcpip.KeepaliveIdleOption(tcpKeepaliveIdle)
	if err := ep.SetSockOpt(&idleOpt); err != nil {
		return fmt.Errorf("set keepalive idle: %s", err)
	}
	intervalOpt := tcpip.KeepaliveIntervalOption(tcpKeepaliveInterval)
	if err := ep.SetSockOpt(&intervalOpt); err != nil {
		return fmt.Errorf("set keepalive interval: %s", err)
	}
	return nil
}

type tcpConn struct {
	net.Conn
	metadata *adapter.Metadata
}

func (c *tcpConn) Metadata() *adapter.Metadata {
	return c.metadata
}
