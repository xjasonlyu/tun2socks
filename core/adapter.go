package core

import (
	"net"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type TCPConn interface {
	net.Conn
	ID() *stack.TransportEndpointID
}

type UDPPacket interface {
	// Data get the payload of UDP Packet.
	Data() []byte

	// Drop call after packet is used, could release resources in this function.
	Drop()

	// ID returns the transport endpoint id of packet.
	ID() *stack.TransportEndpointID

	// LocalAddr returns the source IP/Port of packet.
	LocalAddr() net.Addr

	// RemoteAddr returns the destination IP/Port of packet.
	RemoteAddr() net.Addr

	// WriteBack writes the payload with source IP/Port equals addr
	// - variable source IP/Port is important to STUN
	// - if addr is not provided, WriteBack will write out UDP packet with SourceIP/Port equals to original Target.
	WriteBack([]byte, net.Addr) (int, error)
}
