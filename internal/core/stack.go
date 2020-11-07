package core

import (
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

// NewStack returns *stack.Stack with provided options.
func NewStack(opts ...Option) (*stack.Stack, error) {
	ipstack := stack.New(stack.Options{
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
	})

	for _, opt := range opts {
		if err := opt(ipstack); err != nil {
			return nil, err
		}
	}

	return ipstack, nil
}

// NewDefaultStack calls NewStack with default options.
func NewDefaultStack(linkEp stack.LinkEndpoint, th tcpHandleFunc, uh udpHandleFunc) (*stack.Stack, error) {
	return NewStack(
		WithDefaultTTL(defaultTimeToLive),
		WithForwarding(ipForwardingEnabled),

		// Config default stack ICMP settings.
		WithICMPBurst(icmpBurst), WithICMPLimit(icmpLimit),

		// We expect no packet loss, therefore we can bump buffers.
		// Too large buffers thrash cache, so there is little point
		// in too large buffers.
		//
		// Ref: https://github.com/majek/slirpnetstack/blob/master/stack.go
		WithTCPBufferSizeRange(minBufferSize, defaultBufferSize, maxBufferSize),

		WithTCPCongestionControl(tcpCongestionControlAlgorithm),
		WithTCPDelay(tcpDelayEnabled),

		// Receive Buffer Auto-Tuning Option, see:
		// https://github.com/google/gvisor/issues/1666
		WithTCPModerateReceiveBuffer(tcpModerateReceiveBufferEnabled),

		// TCP selective ACK Option, see:
		// https://tools.ietf.org/html/rfc2018
		WithTCPSACKEnabled(tcpSACKEnabled),

		// Important: We must initiate transport protocol handlers
		// before creating NIC, otherwise NIC would dispatch packets
		// to stack and cause race condition.
		WithICMPHandler(nil), WithTCPHandler(th), WithUDPHandler(uh),

		// Create stack NIC and then bind link endpoint.
		WithCreatingNIC(defaultNICID, linkEp),

		// In past we did s.AddAddressRange to assign 0.0.0.0/0 onto
		// the interface. We need that to be able to terminate all the
		// incoming connections - to any ip. AddressRange API has been
		// removed and the suggested workaround is to use Promiscuous
		// mode. https://github.com/google/gvisor/issues/3876
		//
		// Ref: https://github.com/majek/slirpnetstack/blob/master/stack.go
		WithPromiscuousMode(defaultNICID, nicPromiscuousModeEnabled),

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
		WithSpoofing(defaultNICID, nicSpoofingEnabled),
	)
}
