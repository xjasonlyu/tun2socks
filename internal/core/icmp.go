package core

import (
	"fmt"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type icmpHandleFunc func()

func WithICMPHandler(_ icmpHandleFunc) Option {
	return func(s *stack.Stack) error {
		// Add default route table for IPv4 and IPv6
		// This will handle all incoming ICMP packets.
		s.SetRouteTable([]tcpip.Route{
			{
				Destination: mustSubnet("0.0.0.0/0"),
				NIC:         defaultNICID,
			},
			{
				Destination: mustSubnet("::/0"),
				NIC:         defaultNICID,
			},
		})
		return nil
	}
}

// mustSubnet returns tcpip.Subnet from CIDR string.
func mustSubnet(s string) tcpip.Subnet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Errorf("unable to ParseCIDR(%s): %w", s, err))
	}

	subnet, err := tcpip.NewSubnet(tcpip.Address(ipNet.IP), tcpip.AddressMask(ipNet.Mask))
	if err != nil {
		panic(fmt.Errorf("unable to NewSubnet(%s): %w", ipNet, err))
	}
	return subnet
}
