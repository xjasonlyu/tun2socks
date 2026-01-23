package adapter

import (
	"net"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// TCPConn represents a TCP connection that implements net.Conn
// and exposes its stack.TransportEndpointID.
type TCPConn interface {
	net.Conn

	// ID returns the transport endpoint ID.
	ID() stack.TransportEndpointID
}

// UDPConn represents a UDP connection that implements both net.Conn
// and net.PacketConn and exposes its stack.TransportEndpointID.
type UDPConn interface {
	net.Conn
	net.PacketConn

	// ID returns the transport endpoint ID.
	ID() stack.TransportEndpointID
}

// Packet represents a generic network packet delivered to a network
// handler. It provides access to the underlying packet buffer, the
// owning network stack, and the associated stack.TransportEndpointID.
type Packet interface {
	// Buffer returns the packet buffer containing the data and headers.
	Buffer() *stack.PacketBuffer

	// Stack returns the network stack responsible for handling this packet.
	Stack() *stack.Stack

	// ID returns the transport endpoint ID.
	ID() stack.TransportEndpointID
}
