// Package stack provides a thin wrapper around a gVisor's stack.
package stack

import (
	"github.com/xjasonlyu/tun2socks/v2/core"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

type Stack struct {
	*stack.Stack

	handler core.Handler
	nicID   tcpip.NICID
}

// New allocates a new *Stack with given options.
func New(ep stack.LinkEndpoint, handler core.Handler, opts ...Option) (*Stack, error) {
	s := &Stack{
		Stack: stack.New(stack.Options{
			NetworkProtocols: []stack.NetworkProtocolFactory{
				ipv4.NewProtocol,
				ipv6.NewProtocol,
			},
			TransportProtocols: []stack.TransportProtocolFactory{
				tcp.NewProtocol,
				udp.NewProtocol,
				icmp.NewProtocol4,
				icmp.NewProtocol6,
			},
		}),

		handler: handler,
		nicID:   defaultNICID,
	}

	opts = append(opts,
		// Important: We must initiate transport protocol handlers
		// before creating NIC, otherwise NIC would dispatch packets
		// to stack and cause race condition.
		withICMPHandler(), withTCPHandler(), withUDPHandler(),

		// Create stack NIC and then bind link endpoint.
		withCreatingNIC(ep),

		// In past we did s.AddAddressRange to assign 0.0.0.0/0 onto
		// the interface. We need that to be able to terminate all the
		// incoming connections - to any ip. AddressRange API has been
		// removed and the suggested workaround is to use Promiscuous
		// mode. https://github.com/google/gvisor/issues/3876
		//
		// Ref: https://github.com/majek/slirpnetstack/blob/master/stack.go
		withPromiscuousMode(nicPromiscuousModeEnabled),

		// Enable spoofing if a stack may send packets from unowned addresses.
		// This change required changes to some netgophers since previously,
		// promiscuous mode was enough to let the netstack respond to all
		// incoming packets regardless of the packet's destination address. Now
		// that a stack.Route is not held for each incoming packet, finding a route
		// may fail with local addresses we don't own but accepted packets for
		// while in promiscuous mode. Since we also want to be able to send from
		// any address (in response the received promiscuous mode packets), we need
		// to enable spoofing.
		//
		// Ref: https://github.com/google/gvisor/commit/8c0701462a84ff77e602f1626aec49479c308127
		withSpoofing(nicSpoofingEnabled),
	)

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}
