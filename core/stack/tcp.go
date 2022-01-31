package stack

import (
	"fmt"
	"net"
	"time"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
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

func withTCPHandler() Option {
	return func(s *Stack) error {
		tcpForwarder := tcp.NewForwarder(s.Stack, defaultWndSize, maxConnAttempts, func(r *tcp.ForwarderRequest) {
			var wq waiter.Queue
			id := r.ID()
			ep, err := r.CreateEndpoint(&wq)
			if err != nil {
				// prevent potential half-open TCP connection leak.
				r.Complete(true)
				return
			}
			r.Complete(false)

			setKeepalive(ep)

			conn := &tcpConn{
				Conn: gonet.NewTCPConn(&wq, ep),
				id:   &id,
			}
			s.handler.Add(conn)
		})
		s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)
		return nil
	}
}

func setKeepalive(ep tcpip.Endpoint) error {
	ep.SocketOptions().SetKeepAlive(true)

	idle := tcpip.KeepaliveIdleOption(tcpKeepaliveIdle)
	if err := ep.SetSockOpt(&idle); err != nil {
		return fmt.Errorf("set keepalive idle: %s", err)
	}

	interval := tcpip.KeepaliveIntervalOption(tcpKeepaliveInterval)
	if err := ep.SetSockOpt(&interval); err != nil {
		return fmt.Errorf("set keepalive interval: %s", err)
	}
	return nil
}

type tcpConn struct {
	net.Conn
	id *stack.TransportEndpointID
}

func (c *tcpConn) ID() *stack.TransportEndpointID {
	return c.id
}
