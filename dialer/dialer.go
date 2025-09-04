package dialer

import (
	"context"
	"net"
	"syscall"

	"go.uber.org/atomic"
)

// DefaultDialer is the default Dialer and is used by DialContext and ListenPacket.
var DefaultDialer = &Dialer{
	InterfaceName:  atomic.NewString(""),
	InterfaceIndex: atomic.NewInt32(0),
	RoutingMark:    atomic.NewInt32(0),
}

type Dialer struct {
	InterfaceName  *atomic.String
	InterfaceIndex *atomic.Int32
	RoutingMark    *atomic.Int32
}

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

// DialContext is a wrapper around DefaultDialer.DialContext.
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return DefaultDialer.DialContext(ctx, network, address)
}

// ListenPacket is a wrapper around DefaultDialer.ListenPacket.
func ListenPacket(network, address string) (net.PacketConn, error) {
	return DefaultDialer.ListenPacket(network, address)
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.DialContextWithOptions(ctx, network, address, &Options{
		InterfaceName:  d.InterfaceName.Load(),
		InterfaceIndex: int(d.InterfaceIndex.Load()),
		RoutingMark:    int(d.RoutingMark.Load()),
	})
}

func (*Dialer) DialContextWithOptions(ctx context.Context, network, address string, opts *Options) (net.Conn, error) {
	d := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return setSocketOptions(network, address, c, opts)
		},
	}
	return d.DialContext(ctx, network, address)
}

func (d *Dialer) ListenPacket(network, address string) (net.PacketConn, error) {
	return d.ListenPacketWithOptions(network, address, &Options{
		InterfaceName:  d.InterfaceName.Load(),
		InterfaceIndex: int(d.InterfaceIndex.Load()),
		RoutingMark:    int(d.RoutingMark.Load()),
	})
}

func (*Dialer) ListenPacketWithOptions(network, address string, opts *Options) (net.PacketConn, error) {
	lc := &net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return setSocketOptions(network, address, c, opts)
		},
	}
	return lc.ListenPacket(context.Background(), network, address)
}
