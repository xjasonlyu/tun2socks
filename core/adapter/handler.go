package adapter

// TransportHandler is a TCP/UDP connection handler that implements
// HandleTCPConn and HandleUDPConn methods.
type TransportHandler interface {
	HandleTCP(TCPConn)
	HandleUDP(UDPConn)
}

// TCPHandleFunc handles incoming TCP connection.
type TCPHandleFunc func(TCPConn)

// UDPHandleFunc handles incoming UDP connection.
type UDPHandleFunc func(UDPConn)
