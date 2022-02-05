package stack

import (
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
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

			s.handler.HandleUDPConn(gonet.NewUDPConn(s.Stack, &wq, ep))
		})
		s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)
		return nil
	}
}
