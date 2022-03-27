package core

import (
	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/core/option"

	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func withUDPHandler(handle adapter.UDPHandleFunc) option.Option {
	return func(s *stack.Stack) error {
		udpForwarder := udp.NewForwarder(s, func(r *udp.ForwarderRequest) {
			var wq waiter.Queue
			ep, err := r.CreateEndpoint(&wq)
			if err != nil {
				// TODO: handler errors in the future.
				return
			}

			conn := &udpConn{
				UDPConn: gonet.NewUDPConn(s, &wq, ep),
				id:      r.ID(),
			}
			handle(conn)
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
