package adapter

import "net"

type TCPConn interface {
	net.Conn
	Metadata() *Metadata
}

type UDPPacket interface {
	// Data get the payload of UDP Packet.
	Data() []byte

	// Drop call after packet is used, could release resources in this function.
	Drop()

	// LocalAddr returns the source IP/Port of packet.
	LocalAddr() net.Addr

	// Metadata returns the metadata of packet.
	Metadata() *Metadata

	// RemoteAddr returns the destination IP/Port of packet.
	RemoteAddr() net.Addr

	// WriteBack writes the payload with source IP/Port equals addr
	// - variable source IP/Port is important to STUN
	// - if addr is not provided, WriteBack will write out UDP packet with SourceIP/Port equals to original Target,
	//   this is important when using Fake-IP.
	WriteBack([]byte, net.Addr) (int, error)
}
