package stack

import (
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func withUDPHandler() Option {
	return func(s *Stack) error {
		udpForwarder := udp.NewForwarder(s.Stack, func(r *udp.ForwarderRequest) {
			var wq waiter.Queue
			ep, err := r.CreateEndpoint(&wq)
			if err != nil {
				// TODO: handler errors in the future.
				return
			}

			conn := &udpConn{
				UDPConn: gonet.NewUDPConn(s.Stack, &wq, ep),
				id:      r.ID(),
			}
			s.handler.HandleUDPConn(conn)
		})
		s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)
		return nil
	}
}

type udpConn struct {
	*gonet.UDPConn
	id stack.TransportEndpointID
}

func (c *udpConn) ID() *stack.TransportEndpointID {
	return &c.id
}
