package utils

import (
	"context"
	"fmt"
	"net"
	"time"
)

// DialWithHandshake is used only by the opt-in deferred TCP paths. Connecting
// to the proxy has a fixed 5s budget; the handshake gets a fresh budget after
// that succeeds. Both stages inherit an earlier parent deadline/cancellation.
func DialWithHandshake(ctx context.Context, network, address, protocol string, timeout time.Duration,
	dial func(context.Context, string, string) (net.Conn, error), handshake func(net.Conn) error,
) (net.Conn, error) {
	connectCtx, cancelConnect := context.WithTimeout(ctx, TCPConnectTimeout)
	c, err := dial(connectCtx, network, address)
	if err == nil {
		err = connectCtx.Err()
	}
	cancelConnect()
	if err != nil {
		SafeConnClose(c, err)
		return nil, fmt.Errorf("connect to %s: %w", address, err)
	}
	if c == nil {
		return nil, fmt.Errorf("connect to %s: dialer returned no connection", address)
	}
	SetKeepAlive(c)
	handshakeCtx, cancelHandshake := context.WithTimeout(ctx, timeout)
	defer cancelHandshake()
	err = HandshakeContext(handshakeCtx, c, func() error { return handshake(c) })
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("%s handshake: %w", protocol, err)
	}
	return c, nil
}
