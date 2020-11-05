package core

import (
	"fmt"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const (
	// defaultNICID is the ID of default NIC used by DefaultStack.
	defaultNICID tcpip.NICID = 1

	// nicPromiscuousModeEnabled is the value used by stack to enable
	// or disable NIC's promiscuous mode.
	nicPromiscuousModeEnabled = true

	// nicSpoofingEnabled is the value used by stack to enable or disable
	// NIC's spoofing.
	nicSpoofingEnabled = true
)

// WithCreatingNIC creates NIC for stack.
func WithCreatingNIC(nicID tcpip.NICID, ep stack.LinkEndpoint) Option {
	return func(s *stack.Stack) error {
		if err := s.CreateNIC(nicID, ep); err != nil {
			return fmt.Errorf("create NIC: %s", err)
		}
		return nil
	}
}

// WithPromiscuousMode sets promiscuous mode in the given NIC.
func WithPromiscuousMode(nicID tcpip.NICID, v bool) Option {
	return func(s *stack.Stack) error {
		if err := s.SetPromiscuousMode(nicID, v); err != nil {
			return fmt.Errorf("set promiscuous mode: %s", err)
		}
		return nil
	}
}

// WithSpoofing sets address spoofing in the given NIC, allowing
// endpoints to bind to any address in the NIC.
func WithSpoofing(nicID tcpip.NICID, v bool) Option {
	return func(s *stack.Stack) error {
		if err := s.SetSpoofing(nicID, v); err != nil {
			return fmt.Errorf("set spoofing: %s", err)
		}
		return nil
	}
}
