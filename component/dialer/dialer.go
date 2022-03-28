package dialer

import (
	"context"
	"net"
	"syscall"

	"go.uber.org/atomic"
)

var (
	DefaultInterfaceName = atomic.NewString("")
	DefaultRoutingMark   = atomic.NewInt32(0)
)

type Options struct {
	// InterfaceName is the name of interface/device to bind.
	// If a socket is bound to an interface, only packets received
	// from that particular interface are processed by the socket.
	InterfaceName string

	// RoutingMark is the mark for each packet sent through this
	// socket. Changing the mark can be used for mark-based routing
	// without netfilter or for packet filtering.
	RoutingMark int
}

func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return DialContextWithOptions(ctx, network, address, &Options{
		InterfaceName: DefaultInterfaceName.Load(),
		RoutingMark:   int(DefaultRoutingMark.Load()),
	})
}

func DialContextWithOptions(ctx context.Context, network, address string, opts *Options) (net.Conn, error) {
	d := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return setSocketOptions(network, address, c, opts)
		},
	}
	return d.DialContext(ctx, network, address)
}

func ListenPacket(network, address string) (net.PacketConn, error) {
	return ListenPacketWithOptions(network, address, &Options{
		InterfaceName: DefaultInterfaceName.Load(),
		RoutingMark:   int(DefaultRoutingMark.Load()),
	})
}

func ListenPacketWithOptions(network, address string, opts *Options) (net.PacketConn, error) {
	lc := &net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return setSocketOptions(network, address, c, opts)
		},
	}
	return lc.ListenPacket(context.Background(), network, address)
}
