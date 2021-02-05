package stack

import (
	"fmt"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
)

func withICMPHandler() Option {
	return func(s *Stack) error {
		// Add default route table for IPv4 and IPv6.
		// This will handle all incoming ICMP packets.
		s.SetRouteTable([]tcpip.Route{
			{
				Destination: mustSubnet("0.0.0.0/0"),
				NIC:         s.nicID,
			},
			{
				Destination: mustSubnet("::/0"),
				NIC:         s.nicID,
			},
		})
		return nil
	}
}

// mustSubnet returns tcpip.Subnet from CIDR string.
func mustSubnet(s string) tcpip.Subnet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Errorf("parse CIDR: %w", err))
	}

	subnet, err := tcpip.NewSubnet(tcpip.Address(ipNet.IP), tcpip.AddressMask(ipNet.Mask))
	if err != nil {
		panic(fmt.Errorf("new subnet: %w", err))
	}
	return subnet
}
