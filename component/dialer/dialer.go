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
	setControl(d)
	return d.DialContext(ctx, network, address)
}

func ListenPacket(network, address string) (net.PacketConn, error) {
	lc := &net.ListenConfig{}
	setControl(lc)
	return lc.ListenPacket(context.Background(), network, address)
}
