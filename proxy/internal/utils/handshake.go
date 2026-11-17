package utils

import (
	"context"
	"net"
	"time"
)

// HandshakeContext bounds an opt-in proxy handshake, including cancellation
// without a deadline. Successful relay connections must not retain the timer.
// Legacy proxy DialContext paths intentionally do not call this helper.
func HandshakeContext(ctx context.Context, c net.Conn, handshake func() error) error {
	if deadline, ok := ctx.Deadline(); ok {
		if err := c.SetDeadline(deadline); err != nil {
			return err
		}
	}
	stop := context.AfterFunc(ctx, func() { _ = c.Close() })
	defer stop()
	err := handshake()
	if !stop() || ctx.Err() != nil {
		if cause := ctx.Err(); cause != nil {
			return cause
		}
		return context.Canceled
	}
	if err != nil {
		return err
	}
	return c.SetDeadline(time.Time{})
}
