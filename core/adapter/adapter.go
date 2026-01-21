package adapter

import (
	"net"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// TCPConn represents a TCP connection that implements net.Conn
// and exposes its stack.TransportEndpointID.
type TCPConn interface {
	net.Conn

	// ID returns the transport endpoint id.
	ID() stack.TransportEndpointID
}

// UDPConn represents a UDP connection that implements both net.Conn
// and net.PacketConn and exposes its stack.TransportEndpointID.
type UDPConn interface {
	net.Conn
	net.PacketConn

	// ID returns the transport endpoint id.
	ID() stack.TransportEndpointID
}

type Packet interface {
	// Data holds the packet buffer of Packet.
	Data() *stack.PacketBuffer

	// Stack returns the network stack that handles Packet.
	Stack() *stack.Stack

	// ID returns the transport endpoint id of Packet.
	ID() *stack.TransportEndpointID
}
