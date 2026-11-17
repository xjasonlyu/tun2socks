package tunnel

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type timeoutProxy struct {
	dial func(context.Context, *M.Metadata) (net.Conn, error)
}

func (p timeoutProxy) DialContext(ctx context.Context, m *M.Metadata) (net.Conn, error) {
	return p.dial(ctx, m)
}
func (timeoutProxy) DialUDP(*M.Metadata) (net.PacketConn, error) { return nil, errors.New("unused") }

type stagedTimeoutProxy struct {
	timeoutProxy
	staged func(context.Context, proxy.TCPTimeouts) (net.Conn, error)
}

func (p stagedTimeoutProxy) DialContextWithTimeouts(ctx context.Context, _ *M.Metadata, cfg proxy.TCPTimeouts) (net.Conn, error) {
	return p.staged(ctx, cfg)
}

func TestDeferredHandlerPassesSeparateBudgetsWithoutOldOverallLimit(t *testing.T) {
	cfg := proxy.TCPTimeouts{ProxyHandshakeTimeout: time.Second}
	p := stagedTimeoutProxy{timeoutProxy: timeoutProxy{dial: func(context.Context, *M.Metadata) (net.Conn, error) { t.Fatal("used legacy dial"); return nil, nil }}, staged: func(ctx context.Context, got proxy.TCPTimeouts) (net.Conn, error) {
		if got != cfg {
			t.Fatalf("lost config: %+v", got)
		}
		if _, ok := ctx.Deadline(); ok {
			t.Fatal("overall connect deadline incorrectly caps SOCKS handshake")
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		a, b := net.Pipe()
		t.Cleanup(func() { a.Close(); b.Close() })
		return a, nil
	}}
	tun := New(p, statistic.DefaultManager)
	defer tun.Close()
	h, err := NewDeferredTCPHandlerWithConfig(context.Background(), tun, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	c, err := h.DialTCP(context.Background(), stack.TransportEndpointID{})
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
}

func TestDeferredHandlerStopCancelsSOCKSHandshake(t *testing.T) {
	entered := make(chan struct{})
	p := stagedTimeoutProxy{staged: func(ctx context.Context, _ proxy.TCPTimeouts) (net.Conn, error) {
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	tun := New(p, statistic.DefaultManager)
	defer tun.Close()
	h := NewDeferredTCPHandler(context.Background(), tun)
	defer h.Close()
	done := make(chan error, 1)
	go func() { _, err := h.DialTCP(context.Background(), stack.TransportEndpointID{}); done <- err }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("dial not entered")
	}
	h.Close()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("stop did not cancel handshake")
	}
}

func TestNonSOCKSProxyStillHasFixedConnectLimit(t *testing.T) {
	p := timeoutProxy{dial: func(ctx context.Context, _ *M.Metadata) (net.Conn, error) {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 5*time.Second || time.Until(deadline) < 5*time.Second-200*time.Millisecond {
			t.Error("legacy 5s limit changed")
		}
		return nil, context.DeadlineExceeded
	}}
	tun := New(p, statistic.DefaultManager)
	defer tun.Close()
	h, err := NewDeferredTCPHandlerWithConfig(context.Background(), tun, proxy.TCPTimeouts{})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	started := time.Now()
	_, err = h.DialTCP(context.Background(), stack.TransportEndpointID{})
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > time.Second {
		t.Fatal(err)
	}
}

func TestDeferredHandlerRejectsNegativeTimeout(t *testing.T) {
	if h, err := NewDeferredTCPHandlerWithConfig(context.Background(), nil, proxy.TCPTimeouts{ProxyHandshakeTimeout: -1}); err == nil || h != nil {
		t.Fatal("invalid timeout accepted")
	}
}
