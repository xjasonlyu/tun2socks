package adapter

// TransportHandler is a TCP/UDP connection handler that implements
// HandleTCP and HandleUDP methods.
type TransportHandler interface {
	HandleTCP(TCPConn)
	HandleUDP(UDPConn)
}

// NetworkHandler is a L3/network packet handler that implements
// HandlePacket method.
type NetworkHandler interface {
	HandlePacket(Packet) bool
}
