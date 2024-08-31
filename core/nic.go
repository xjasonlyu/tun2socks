package core

import (
	"fmt"
	"net/netip"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core/option"
)

const (
	// nicPromiscuousModeEnabled is the value used by stack to enable
	// or disable NIC's promiscuous mode.
	nicPromiscuousModeEnabled = true

	// nicSpoofingEnabled is the value used by stack to enable or disable
	// NIC's spoofing.
	nicSpoofingEnabled = true
)

// withCreatingNIC creates NIC for stack.
func withCreatingNIC(nicID tcpip.NICID, ep stack.LinkEndpoint) option.Option {
	return func(s *stack.Stack) error {
		if err := s.CreateNICWithOptions(nicID, ep,
			stack.NICOptions{
				Disabled: false,
				// If no queueing discipline was specified
				// provide a stub implementation that just
				// delegates to the lower link endpoint.
				QDisc: nil,
			}); err != nil {
			return fmt.Errorf("create NIC: %s", err)
		}
		return nil
	}
}

// withPromiscuousMode sets promiscuous mode in the given NICs.
func withPromiscuousMode(nicID tcpip.NICID, v bool) option.Option {
	return func(s *stack.Stack) error {
		if err := s.SetPromiscuousMode(nicID, v); err != nil {
			return fmt.Errorf("set promiscuous mode: %s", err)
		}
		return nil
	}
}

// withSpoofing sets address spoofing in the given NICs, allowing
// endpoints to bind to any address in the NIC.
func withSpoofing(nicID tcpip.NICID, v bool) option.Option {
	return func(s *stack.Stack) error {
		if err := s.SetSpoofing(nicID, v); err != nil {
			return fmt.Errorf("set spoofing: %s", err)
		}
		return nil
	}
}

// withMulticastGroups adds a NIC to the given multicast groups.
func withMulticastGroups(nicID tcpip.NICID, multicastGroups []netip.Addr) option.Option {
	return func(s *stack.Stack) error {
		if len(multicastGroups) == 0 {
			return nil
		}
		// The default NIC of tun2socks is working on Spoofing mode. When the UDP Endpoint
		// tries to use a non-local address to connect, the network stack will
		// generate a temporary addressState to build the route, which can be primary
		// but is ephemeral. Nevertheless, when the UDP Endpoint tries to use a
		// multicast address to connect, the network stack will select an available
		// primary addressState to build the route. However, when tun2socks is in the
		// just-initialized or idle state, there will be no available primary addressState,
		// and the connect operation will fail. Therefore, we need to add permanent addresses,
		// e.g. 10.0.0.1/8 and fd00:1/8, to the default NIC, which are only used to build
		// routes for multicast response and do not affect other connections.
		//
		// In fact, for multicast, the sender normally does not expect a response.
		// So, the ep.net.Connect is unnecessary. If we implement a custom UDP Forwarder
		// and ForwarderRequest in the future, we can remove these code.
		s.AddProtocolAddress(
			nicID,
			tcpip.ProtocolAddress{
				Protocol: ipv4.ProtocolNumber,
				AddressWithPrefix: tcpip.AddressWithPrefix{
					Address:   tcpip.AddrFrom4([4]byte{0x0a, 0, 0, 0x01}),
					PrefixLen: 8,
				},
			},
			stack.AddressProperties{PEB: stack.CanBePrimaryEndpoint},
		)
		s.AddProtocolAddress(
			nicID,
			tcpip.ProtocolAddress{
				Protocol: ipv6.ProtocolNumber,
				AddressWithPrefix: tcpip.AddressWithPrefix{
					Address:   tcpip.AddrFrom16([16]byte{0xfd, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}),
					PrefixLen: 8,
				},
			},
			stack.AddressProperties{PEB: stack.CanBePrimaryEndpoint},
		)
		for _, multicastGroup := range multicastGroups {
			var err tcpip.Error
			switch {
			case multicastGroup.Is4():
				err = s.JoinGroup(ipv4.ProtocolNumber, nicID, tcpip.AddrFrom4(multicastGroup.As4()))
			case multicastGroup.Is6():
				err = s.JoinGroup(ipv6.ProtocolNumber, nicID, tcpip.AddrFrom16(multicastGroup.As16()))
			}
			if err != nil {
				return fmt.Errorf("join multicast group: %s", err)
			}
		}
		return nil
	}
}
