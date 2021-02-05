package dialer

import (
	"context"
	"net"
)

func Dial(network, address string) (net.Conn, error) {
	return DialContext(context.Background(), network, address)
}

func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d := &net.Dialer{}

	if _boundInterface != nil {
		d.Control = bindToInterface
	}

	return d.DialContext(ctx, network, address)
}

func ListenPacket(network, address string) (net.PacketConn, error) {
	lc := &net.ListenConfig{}

	if _boundInterface != nil {
		lc.Control = bindToInterface
	}

	return lc.ListenPacket(context.Background(), network, address)
}
