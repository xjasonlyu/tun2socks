package adapter

import (
	"net"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// TCPConnSYN
type TCPConnSYN interface {
	// ID returns the transport endpoint id of TCPConn.
	ID() *stack.TransportEndpointID

	CompleteHandshake() (net.Conn, error)
	StopHandshake()
}

// UDPConn implements net.Conn and net.PacketConn.
type UDPConn interface {
	net.Conn
	net.PacketConn

	// ID returns the transport endpoint id of UDPConn.
	ID() *stack.TransportEndpointID
}
