package adapter

import (
	"context"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// TransportHandler is a TCP/UDP connection handler that implements
// HandleTCP and HandleUDP methods.
type TransportHandler interface {
	HandleTCP(TCPConn)
	HandleUDP(UDPConn)
}

// TCPConnector is used only when core.Config.DeferredTCP is configured. Merely
// implementing it does not change core.CreateStack's legacy behavior.
// DialTCP must honor ctx, including any proxy handshake, and must
// not return success merely because the proxy's listening socket is connected.
// HandleTCPConnected takes ownership of both connections; it must not dial again.
// The existing TransportHandler interface is unchanged.
type TCPConnector interface {
	DialTCP(context.Context, stack.TransportEndpointID) (net.Conn, error)
	HandleTCPConnected(TCPConn, net.Conn)
}

// NetworkHandler is a L3/network packet handler that implements
// HandlePacket method.
type NetworkHandler interface {
	HandlePacket(Packet) bool
}
