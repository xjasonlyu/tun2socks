package device

import (
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// Device is the interface that implemented by network layer devices (e.g. tun),
// and easy to use as stack.LinkEndpoint.
type Device interface {
	stack.LinkEndpoint

	// Close stops and closes the device.
	Close() error

	// Name returns the current name of the device.
	Name() string

	// Type returns the driver type of the device.
	Type() string
}
