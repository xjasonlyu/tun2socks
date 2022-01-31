package stack

import (
	"fmt"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const (
	// defaultNICID is the ID of default NIC used by DefaultStack.
	defaultNICID tcpip.NICID = 0x01

	// nicPromiscuousModeEnabled is the value used by stack to enable
	// or disable NIC's promiscuous mode.
	nicPromiscuousModeEnabled = true

	// nicSpoofingEnabled is the value used by stack to enable or disable
	// NIC's spoofing.
	nicSpoofingEnabled = true
)

// withCreatingNIC creates NIC for stack.
func withCreatingNIC(ep stack.LinkEndpoint) Option {
	return func(s *Stack) error {
		if err := s.CreateNIC(s.nicID, ep); err != nil {
			return fmt.Errorf("create NIC: %s", err)
		}
		return nil
	}
}

// withPromiscuousMode sets promiscuous mode in the given NIC.
func withPromiscuousMode(v bool) Option {
	return func(s *Stack) error {
		if err := s.SetPromiscuousMode(s.nicID, v); err != nil {
			return fmt.Errorf("set promiscuous mode: %s", err)
		}
		return nil
	}
}

// withSpoofing sets address spoofing in the given NIC, allowing
// endpoints to bind to any address in the NIC.
func withSpoofing(v bool) Option {
	return func(s *Stack) error {
		if err := s.SetSpoofing(s.nicID, v); err != nil {
			return fmt.Errorf("set spoofing: %s", err)
		}
		return nil
	}
}
