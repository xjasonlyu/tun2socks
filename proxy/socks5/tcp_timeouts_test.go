package socks5

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/proxy"
)

func TestProxyConnectTimeoutIsSeparate(t *testing.T) {
	p, _ := New("127.0.0.1:1080", "", "")
	started := time.Now()
	parent, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err := p.dialContextWithTimeouts(parent, testMetadata(), proxy.TCPTimeouts{ProxyHandshakeTimeout: time.Second}, func(ctx context.Context, _, _ string) (net.Conn, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "connect to") || time.Since(started) > time.Second {
		t.Fatalf("connect timeout: %v", err)
	}
}

type deadlineObserver struct {
	net.Conn
	set func(time.Time)
}

func (c *deadlineObserver) SetDeadline(d time.Time) error { c.set(d); return c.Conn.SetDeadline(d) }

func TestHandshakeGetsFreshBudgetAfterProxyConnect(t *testing.T) {
	for _, cfg := range []proxy.TCPTimeouts{{}, {ProxyHandshakeTimeout: 7 * time.Second}} {
		addr, _, closed := testSOCKSServer(t, "success", 0)
		p, _ := New(addr, "", "")
		want, _ := cfg.WithDefaults()
		var dialCtx context.Context
		var connectedAt time.Time
		var handshakeSet, cleared bool
		c, err := p.dialContextWithTimeouts(context.Background(), testMetadata(), cfg, func(ctx context.Context, network, address string) (net.Conn, error) {
			dialCtx = ctx
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > 5*time.Second || time.Until(deadline) < 5*time.Second-200*time.Millisecond {
				t.Fatal("wrong connect budget")
			}
			conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			time.Sleep(40 * time.Millisecond) // Simulate time spent in proxy connect.
			connectedAt = time.Now()
			return &deadlineObserver{Conn: conn, set: func(d time.Time) {
				if d.IsZero() {
					cleared = true
					return
				}
				handshakeSet = true
				if dialCtx.Err() != context.Canceled {
					t.Error("proxy stage context not released")
				}
				budget := d.Sub(connectedAt)
				if budget < want.ProxyHandshakeTimeout || budget > want.ProxyHandshakeTimeout+200*time.Millisecond {
					t.Errorf("handshake reused proxy deadline: %v", budget)
				}
			}}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
		waitClosed(t, closed)
		if !handshakeSet || !cleared {
			t.Fatal("handshake deadline not set/cleared")
		}
	}
}

func TestProxyHandshakeTimeoutCoversAllStages(t *testing.T) {
	for _, stage := range []string{"greeting", "authentication", "connect"} {
		t.Run(stage, func(t *testing.T) {
			addr, reached, closed := testSOCKSServer(t, stage, 0)
			user, pass := "", ""
			if stage == "authentication" {
				user, pass = "user", "pass"
			}
			p, _ := New(addr, user, pass)
			started := time.Now()
			c, err := p.DialContextWithTimeouts(context.Background(), testMetadata(), proxy.TCPTimeouts{ProxyHandshakeTimeout: 80 * time.Millisecond})
			if c != nil {
				c.Close()
			}
			if err == nil || !strings.Contains(err.Error(), "SOCKS5 handshake") || time.Since(started) > time.Second {
				t.Fatalf("stage %s timeout failed: %v", stage, err)
			}
			select {
			case <-reached:
			default:
				t.Fatal("stage never reached")
			}
			waitClosed(t, closed)
		})
	}
}

func TestParentDeadlineCapsSOCKSHandshakeBudget(t *testing.T) {
	addr, _, closed := testSOCKSServer(t, "connect", 0)
	p, _ := New(addr, "", "")
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	started := time.Now()
	c, err := p.DialContextWithTimeouts(ctx, testMetadata(), proxy.TCPTimeouts{ProxyHandshakeTimeout: time.Second})
	if c != nil {
		c.Close()
	}
	if err == nil || time.Since(started) > 500*time.Millisecond {
		t.Fatal("parent deadline ignored", err)
	}
	waitClosed(t, closed)
}

func TestConfiguredHandshakeDeadlineDoesNotLimitRelay(t *testing.T) {
	addr, _, closed := testSOCKSServer(t, "success", 0)
	p, _ := New(addr, "", "")
	c, err := p.DialContextWithTimeouts(context.Background(), testMetadata(), proxy.TCPTimeouts{ProxyHandshakeTimeout: 80 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	time.Sleep(120 * time.Millisecond)
	if _, err = c.Write([]byte("echo")); err != nil {
		t.Fatal(err)
	}
	_ = c.SetReadDeadline(time.Now().Add(time.Second))
	b := make([]byte, 4)
	if _, err = io.ReadFull(c, b); err != nil || string(b) != "echo" {
		t.Fatal(string(b), err)
	}
	c.Close()
	waitClosed(t, closed)
}

func TestInvalidTimeoutRejectedBeforeDial(t *testing.T) {
	p, _ := New("127.0.0.1:1080", "", "")
	for _, cfg := range []proxy.TCPTimeouts{{ProxyHandshakeTimeout: -1}} {
		_, err := p.dialContextWithTimeouts(context.Background(), testMetadata(), cfg, func(context.Context, string, string) (net.Conn, error) {
			t.Fatal("dialed with invalid config")
			return nil, nil
		})
		if err == nil {
			t.Fatal("accepted invalid timeout")
		}
	}
}
