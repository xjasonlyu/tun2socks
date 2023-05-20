package dialer

import (
	"context"
	"net"
	"syscall"

	"go.uber.org/atomic"
)

var (
	DefaultInterfaceName  = atomic.NewString("")
	DefaultInterfaceIndex = atomic.NewInt32(0)
	DefaultRoutingMark    = atomic.NewInt32(0)
)

type Options struct {
	// InterfaceName is the name of interface/device to bind.
	// If a socket is bound to an interface, only packets received
	// from that particular interface are processed by the socket.
	InterfaceName string

	// InterfaceIndex is the index of interface/device to bind.
	// It is almost the same as InterfaceName except it uses the
	// index of the interface instead of the name.
	InterfaceIndex int

	// RoutingMark is the mark for each packet sent through this
	// socket. Changing the mark can be used for mark-based routing
	// without netfilter or for packet filtering.
	RoutingMark int
}

func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return DialContextWithOptions(ctx, network, address, &Options{
		InterfaceName:  DefaultInterfaceName.Load(),
		InterfaceIndex: int(DefaultInterfaceIndex.Load()),
		RoutingMark:    int(DefaultRoutingMark.Load()),
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
		InterfaceName:  DefaultInterfaceName.Load(),
		InterfaceIndex: int(DefaultInterfaceIndex.Load()),
		RoutingMark:    int(DefaultRoutingMark.Load()),
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
