package tunnel

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"testing"
	"time"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	httpproxy "github.com/xjasonlyu/tun2socks/v2/proxy/http"
	"github.com/xjasonlyu/tun2socks/v2/proxy/socks4"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func realHandshakeProxy(t *testing.T, protocol string, succeed bool, replyGate <-chan struct{}) (proxy.Proxy, <-chan struct{}, <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	entered, closed := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(closed)
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_ = c.SetDeadline(time.Now().Add(10 * time.Second))
		reader := bufio.NewReader(c)
		if protocol == "http" {
			if _, err := http.ReadRequest(reader); err != nil {
				return
			}
		} else {
			var head [8]byte
			if _, err := io.ReadFull(reader, head[:]); err != nil {
				return
			}
			if _, err := reader.ReadBytes(0); err != nil {
				return
			}
		}
		close(entered)
		if succeed {
			if replyGate != nil {
				select {
				case <-replyGate:
				case <-time.After(5 * time.Second):
					return
				}
			}
			if protocol == "http" {
				_, _ = io.WriteString(c, "HTTP/1.1 200 Connection established\r\n\r\n")
			} else {
				_, _ = c.Write([]byte{0, 90, 0, 0, 0, 0, 0, 0})
			}
			_, _ = io.Copy(c, reader)
		} else {
			_, _ = io.Copy(io.Discard, reader) // Never send a handshake reply.
		}
	}()
	var p proxy.Proxy
	if protocol == "http" {
		p, err = httpproxy.New(ln.Addr().String(), "", "")
	} else {
		p, err = socks4.New(ln.Addr().String(), "")
	}
	if err != nil {
		t.Fatal(err)
	}
	return p, entered, closed
}

func TestDeferredRealProxyStalledHandshake(t *testing.T) {
	for _, protocol := range []string{"http", "socks4"} {
		for _, mode := range []string{"deadline", "parent-cancel", "handler-close", "configured-timeout"} {
			t.Run(protocol+"/"+mode, func(t *testing.T) {
				t.Parallel()
				p, entered, closed := realHandshakeProxy(t, protocol, false, nil)
				tun := New(p, statistic.DefaultManager)
				cfg := proxy.TCPTimeouts{}
				if mode == "configured-timeout" {
					cfg.ProxyHandshakeTimeout = 100 * time.Millisecond
				}
				h, err := NewDeferredTCPHandlerWithConfig(context.Background(), tun, cfg)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { h.Close(); tun.Close() })
				ctx, cancel := context.WithCancel(context.Background())
				if mode == "deadline" {
					cancel()
					ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
				}
				defer cancel()
				done := make(chan error, 1)
				start := time.Now()
				go func() {
					c, err := h.DialTCP(ctx, stack.TransportEndpointID{LocalAddress: tcpip.AddrFrom4([4]byte{1, 2, 3, 4}), LocalPort: 443})
					if c != nil {
						c.Close()
					}
					done <- err
				}()
				select {
				case <-entered:
				case <-time.After(2 * time.Second):
					t.Fatal("handshake not entered")
				}
				if mode == "parent-cancel" {
					cancel()
				}
				if mode == "handler-close" {
					h.Close()
				}
				limit := 2 * time.Second
				select {
				case err := <-done:
					if err == nil {
						t.Fatal("stalled proxy succeeded")
					}
					if (mode == "handler-close" || mode == "parent-cancel") && !errors.Is(err, context.Canceled) {
						t.Fatalf("lost cancellation: %v", err)
					}
					if mode == "configured-timeout" && time.Since(start) < cfg.ProxyHandshakeTimeout {
						t.Fatal("handshake ended before configured timeout")
					}
				case <-time.After(limit):
					t.Fatal("dial remained blocked")
				}
				select {
				case <-closed:
				case <-time.After(time.Second):
					t.Fatal("proxy socket leaked")
				}
			})
		}
	}
}

func TestDeferredRealProxySuccessClearsHandshakeDeadline(t *testing.T) {
	for _, protocol := range []string{"http", "socks4"} {
		t.Run(protocol, func(t *testing.T) {
			t.Parallel()
			p, _, closed := realHandshakeProxy(t, protocol, true, nil)
			tun := New(p, statistic.DefaultManager)
			h, err := NewDeferredTCPHandlerWithConfig(context.Background(), tun, proxy.TCPTimeouts{ProxyHandshakeTimeout: 100 * time.Millisecond})
			if err != nil {
				t.Fatal(err)
			}
			defer h.Close()
			defer tun.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			c, err := h.DialTCP(ctx, stack.TransportEndpointID{LocalAddress: tcpip.AddrFrom4([4]byte{1, 2, 3, 4}), LocalPort: 443})
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			<-ctx.Done()
			if _, err := c.Write([]byte("ok")); err != nil {
				t.Fatal("handshake deadline leaked into relay:", err)
			}
			_ = c.SetReadDeadline(time.Now().Add(time.Second))
			var b [2]byte
			if _, err := io.ReadFull(c, b[:]); err != nil || string(b[:]) != "ok" {
				t.Fatal(string(b[:]), err)
			}
			c.Close()
			select {
			case <-closed:
			case <-time.After(time.Second):
				t.Fatal("server did not finish")
			}
		})
	}
}

// The opt-in fix must not silently change legacy DialContext semantics.
func TestLegacyHTTPAndSOCKS4HandshakeUnchanged(t *testing.T) {
	for _, protocol := range []string{"http", "socks4"} {
		t.Run(protocol, func(t *testing.T) {
			t.Parallel()
			gate := make(chan struct{})
			p, entered, _ := realHandshakeProxy(t, protocol, true, gate)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			type result struct {
				c   net.Conn
				err error
			}
			done := make(chan result, 1)
			go func() {
				c, err := p.DialContext(ctx, &M.Metadata{Network: M.TCP, DstIP: netip.MustParseAddr("1.2.3.4"), DstPort: 443})
				done <- result{c, err}
			}()
			select {
			case <-entered:
			case <-time.After(time.Second):
				close(gate)
				t.Fatal("handshake not entered")
			}
			cancel()
			select {
			case r := <-done:
				if r.c != nil {
					r.c.Close()
				}
				close(gate)
				t.Fatal("legacy handshake semantics changed", r.err)
			case <-time.After(30 * time.Millisecond):
			}
			close(gate)
			select {
			case r := <-done:
				if r.c != nil {
					r.c.Close()
				}
				if r.err != nil {
					t.Fatal(r.err)
				}
			case <-time.After(time.Second):
				t.Fatal("legacy handshake did not complete")
			}
		})
	}
}
