package utils

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

type observedConn struct {
	net.Conn
	deadline func(time.Time)
}

func (c *observedConn) SetDeadline(d time.Time) error { c.deadline(d); return c.Conn.SetDeadline(d) }

func TestStagedConnectFreshHandshakeBudget(t *testing.T) {
	for _, duration := range []time.Duration{21 * time.Second, 75 * time.Millisecond} {
		t.Run(duration.String(), func(t *testing.T) {
			client, server := net.Pipe()
			defer server.Close()
			defer client.Close()
			var connectCtx context.Context
			var connected time.Time
			set, cleared := false, false
			c, err := DialWithHandshake(context.Background(), "tcp", "proxy", "test", duration,
				func(ctx context.Context, _, _ string) (net.Conn, error) {
					connectCtx = ctx
					d, ok := ctx.Deadline()
					if !ok || time.Until(d) > 5*time.Second || time.Until(d) < 4500*time.Millisecond {
						t.Fatal("connect budget is not 5s")
					}
					time.Sleep(30 * time.Millisecond)
					connected = time.Now()
					return &observedConn{client, func(d time.Time) {
						if d.IsZero() {
							cleared = true
							return
						}
						set = true
						if connectCtx.Err() != context.Canceled {
							t.Error("connect context not released")
						}
						if budget := d.Sub(connected); budget < duration || budget > duration+500*time.Millisecond {
							t.Errorf("handshake budget %v, want fresh %v", budget, duration)
						}
					}}, nil
				}, func(net.Conn) error { return nil })
			if err != nil {
				t.Fatal(err)
			}
			c.Close()
			if !set || !cleared {
				t.Fatal("handshake deadline not set/cleared")
			}
		})
	}
}

func TestStagedConnectParentWinsDuringDial(t *testing.T) {
	for _, mode := range []string{"deadline", "cancel", "late-success"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
			defer cancel()
			client, server := net.Pipe()
			defer client.Close()
			defer server.Close()
			c, err := DialWithHandshake(ctx, "tcp", "proxy", "test", 21*time.Second,
				func(dialCtx context.Context, _, _ string) (net.Conn, error) {
					if mode != "deadline" {
						cancel()
					}
					<-dialCtx.Done()
					if mode == "late-success" {
						return client, nil
					}
					return client, dialCtx.Err()
				}, func(net.Conn) error { t.Fatal("handshake started after canceled dial"); return nil })
			if c != nil || !errors.Is(err, ctx.Err()) {
				t.Fatal(c, err)
			}
			if _, err := server.Write([]byte("x")); !errors.Is(err, io.ErrClosedPipe) {
				t.Fatal("failed dial connection leaked", err)
			}
		})
	}
}

func TestStagedConnectParentWinsDuringHandshake(t *testing.T) {
	for _, mode := range []string{"deadline", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			client, server := net.Pipe()
			defer server.Close()
			defer client.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			start := time.Now()
			c, err := DialWithHandshake(ctx, "tcp", "proxy", "test", 21*time.Second,
				func(context.Context, string, string) (net.Conn, error) { return client, nil },
				func(c net.Conn) error {
					if mode == "cancel" {
						cancel()
					}
					var b [1]byte
					_, err := c.Read(b[:])
					return err
				})
			if c != nil || err == nil || time.Since(start) > time.Second {
				t.Fatal("parent did not bound handshake", err)
			}
			if mode == "cancel" && !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			if _, err := server.Write([]byte("x")); !errors.Is(err, io.ErrClosedPipe) {
				t.Fatal("handshake connection leaked", err)
			}
		})
	}
}
