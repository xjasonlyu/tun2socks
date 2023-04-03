package adapter

import (
	"net"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// TCPConn implements the net.Conn interface.
type TCPConn interface {
	net.Conn

	// ID returns the transport endpoint id of TCPConn.
	ID() *stack.TransportEndpointID
}

// UDPConn implements net.Conn and net.PacketConn.
type UDPConn interface {
	net.Conn
	net.PacketConn

	// ID returns the transport endpoint id of UDPConn.
	ID() *stack.TransportEndpointID
}

// CloseReader shuts down the reading side of the TCP connection.
type CloseReader interface {
	CloseRead() error
}

// CloseWriter shuts down the writing side of the TCP connection.
type CloseWriter interface {
	CloseWrite() error
}

// DuplexConn implements the net.Conn interface and keeps the
// ability to close reading/writing side of the TCP connection.
type DuplexConn interface {
	net.Conn
	CloseReader
	CloseWriter
}
