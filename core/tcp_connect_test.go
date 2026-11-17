package core_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/core"
	M "github.com/xjasonlyu/tun2socks/v2/metadata"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
)

type testProxy struct {
	dial func(context.Context, *M.Metadata) (net.Conn, error)
}

func (p testProxy) DialContext(ctx context.Context, m *M.Metadata) (net.Conn, error) {
	return p.dial(ctx, m)
}
func (p testProxy) DialUDP(*M.Metadata) (net.PacketConn, error) { return nil, errors.New("unused") }

type tcpHarness struct {
	client   *stack.Stack
	tunnel   *tunnel.Tunnel
	deferred *tunnel.DeferredTCPHandler
	cancel   context.CancelFunc
	synACK   atomic.Int32
	resets   atomic.Int32
}

func tcpAddr(ip string) tcpip.Address { return tcpip.AddrFromSlice(netip.MustParseAddr(ip).AsSlice()) }

// Two real gVisor stacks exchange raw IP packets through an in-memory wire. No
// host routes, elevated privileges, external servers or physical IPv6 required.
func newTCPHarness(t *testing.T, dial func(context.Context, *M.Metadata) (net.Conn, error)) *tcpHarness {
	return newTCPModeHarness(t, dial, true)
}

func newTCPModeHarness(t *testing.T, dial func(context.Context, *M.Metadata) (net.Conn, error), deferred bool) *tcpHarness {
	return newTCPPolicyHarness(t, dial, deferred, false)
}

func newTCPPolicyHarness(t *testing.T, dial func(context.Context, *M.Metadata) (net.Conn, error), deferred, rejectIPv6 bool) *tcpHarness {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	wireCtx, stopWire := context.WithCancel(context.Background())
	clientLink, serverLink := channel.New(1024, 1500, ""), channel.New(1024, 1500, "")
	h := &tcpHarness{cancel: cancel}
	h.tunnel = tunnel.New(testProxy{dial}, statistic.DefaultManager)
	h.tunnel.ProcessAsync()
	h.deferred = tunnel.NewDeferredTCPHandler(ctx, h.tunnel)
	cfg := &core.Config{LinkEndpoint: serverLink, TransportHandler: h.deferred, RejectIPv6: rejectIPv6}
	if deferred {
		cfg.DeferredTCP = &core.DeferredTCPConfig{Context: ctx}
	}
	// Even a TCPConnector must NOT silently enable the new behavior.
	server, err := core.CreateStack(cfg)
	if err != nil {
		t.Fatal(err)
	}
	h.client = stack.New(stack.Options{
		NetworkProtocols:   []stack.NetworkProtocolFactory{ipv4.NewProtocol, ipv6.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{tcp.NewProtocol},
	})
	if err := h.client.CreateNIC(1, clientLink); err != nil {
		t.Fatal(err)
	}
	for _, ip := range []string{"10.0.0.2", "fd00::2"} {
		proto, prefix := ipv4.ProtocolNumber, 24
		if netip.MustParseAddr(ip).Is6() {
			proto, prefix = ipv6.ProtocolNumber, 64
		}
		if err := h.client.AddProtocolAddress(1, tcpip.ProtocolAddress{Protocol: proto, AddressWithPrefix: tcpip.AddressWithPrefix{Address: tcpAddr(ip), PrefixLen: prefix}}, stack.AddressProperties{}); err != nil {
			t.Fatal(err)
		}
	}
	h.client.SetRouteTable([]tcpip.Route{{Destination: header.IPv4EmptySubnet, NIC: 1}, {Destination: header.IPv6EmptySubnet, NIC: 1}})
	var pumps sync.WaitGroup
	pump := func(src, dst *channel.Endpoint, observe bool) {
		defer pumps.Done()
		for {
			pkt := src.ReadContext(wireCtx)
			if pkt == nil {
				return
			}
			if observe {
				// Reject replies may store the TCP header in payload rather than
				// PacketBuffer.TransportHeader. Observe the actual wire bytes.
				buf := pkt.ToBuffer()
				wire := buf.Flatten()
				var tcpBytes []byte
				if pkt.NetworkProtocolNumber == ipv6.ProtocolNumber && header.IPv6(wire).TransportProtocol() == tcp.ProtocolNumber {
					tcpBytes = wire[header.IPv6MinimumSize:]
				} else if pkt.NetworkProtocolNumber == ipv4.ProtocolNumber && header.IPv4(wire).TransportProtocol() == tcp.ProtocolNumber {
					tcpBytes = wire[header.IPv4(wire).HeaderLength():]
				}
				if len(tcpBytes) >= header.TCPMinimumSize {
					flags := header.TCP(tcpBytes).Flags()
					if flags.Contains(header.TCPFlagSyn | header.TCPFlagAck) {
						h.synACK.Add(1)
					}
					if flags.Contains(header.TCPFlagRst) {
						h.resets.Add(1)
					}
				}
				buf.Release()
			}
			incoming := stack.NewPacketBuffer(stack.PacketBufferOptions{Payload: pkt.ToBuffer()})
			dst.InjectInbound(pkt.NetworkProtocolNumber, incoming)
			incoming.DecRef()
			pkt.DecRef()
		}
	}
	pumps.Add(2)
	go pump(clientLink, serverLink, false)
	go pump(serverLink, clientLink, true)
	t.Cleanup(func() {
		cancel()
		h.deferred.Close()
		h.tunnel.Close()
		server.Close()
		h.client.Close()
		server.Wait()
		h.client.Wait()
		stopWire()
		clientLink.Close()
		serverLink.Close()
		pumps.Wait()
	})
	return h
}

func (h *tcpHarness) dial(ip string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	proto := ipv4.ProtocolNumber
	if netip.MustParseAddr(ip).Is6() {
		proto = ipv6.ProtocolNumber
	}
	c, err := gonet.DialContextTCP(ctx, h.client, tcpip.FullAddress{NIC: 1, Addr: tcpAddr(ip), Port: 443}, proto)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func echoRemote() net.Conn {
	a, b := net.Pipe()
	go func() { defer b.Close(); _, _ = io.Copy(b, b) }()
	return a
}

func assertEcho(t *testing.T, c net.Conn) {
	t.Helper()
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(time.Second))
	if _, err := c.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 5)
	if _, err := io.ReadFull(c, buf); err != nil || string(buf) != "hello" {
		t.Fatalf("echo=%q err=%v", buf, err)
	}
}

func TestTCPDoesNotSendSYNACKBeforeRemoteConnect(t *testing.T) {
	for _, ip := range []string{"192.0.2.10", "2001:db8::10"} {
		t.Run(ip, func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			h := newTCPHarness(t, func(ctx context.Context, m *M.Metadata) (net.Conn, error) {
				close(entered)
				select {
				case <-release:
					return echoRemote(), nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			})
			type result struct {
				c   net.Conn
				err error
			}
			done := make(chan result, 1)
			go func() { c, err := h.dial(ip); done <- result{c, err} }()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("outbound dial not started")
			}
			select {
			case result := <-done:
				if result.c != nil {
					result.c.Close()
				}
				t.Fatalf("connect completed before remote: %v", result.err)
			case <-time.After(100 * time.Millisecond):
			}
			if h.synACK.Load() != 0 {
				t.Fatal("premature SYN-ACK")
			}
			close(release)
			outcome := <-done
			if outcome.err != nil {
				t.Fatal(outcome.err)
			}
			assertEcho(t, outcome.c)
			if h.synACK.Load() == 0 {
				t.Fatal("no SYN-ACK observed after success")
			}
		})
	}
}

func TestDefaultModeStillCompletesHandshakeBeforeRemoteConnect(t *testing.T) {
	for _, ip := range []string{"192.0.2.10", "2001:db8::10"} {
		t.Run(ip, func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			defer once.Do(func() { close(release) })
			h := newTCPModeHarness(t, func(ctx context.Context, _ *M.Metadata) (net.Conn, error) {
				close(entered)
				select {
				case <-release:
					return echoRemote(), nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}, false)
			c, err := h.dial(ip)
			if err != nil {
				t.Fatalf("legacy local handshake waited for remote: %v", err)
			}
			defer c.Close()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("legacy handler did not dial")
			}
			if h.synACK.Load() == 0 {
				t.Fatal("legacy mode did not send SYN-ACK before remote success")
			}
			once.Do(func() { close(release) })
			assertEcho(t, c)
		})
	}
}

func TestDefaultModeRemoteFailureDoesNotBecomeHandshakeRejection(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	h := newTCPModeHarness(t, func(ctx context.Context, _ *M.Metadata) (net.Conn, error) {
		select {
		case <-release:
		case <-ctx.Done():
		}
		return nil, errors.New("remote failure")
	}, false)
	c, err := h.dial("192.0.2.10")
	if err != nil {
		t.Fatalf("default mode no longer has early handshake: %v", err)
	}
	c.Close()
	if h.synACK.Load() == 0 {
		t.Fatal("default handshake missing")
	}
}

func TestTCPRejectsSYNInsteadOfReturningAConnectedSocket(t *testing.T) {
	for _, ip := range []string{"192.0.2.10", "2001:db8::10"} {
		t.Run(ip, func(t *testing.T) {
			h := newTCPHarness(t, func(context.Context, *M.Metadata) (net.Conn, error) { return nil, errors.New("remote unreachable") })
			c, err := h.dial(ip)
			if c != nil {
				c.Close()
			}
			if err == nil || h.synACK.Load() != 0 || h.resets.Load() == 0 {
				t.Fatalf("err=%v SYNACK=%d RST=%d", err, h.synACK.Load(), h.resets.Load())
			}
		})
	}
}

func TestIPv6ConnectFailureAllowsIPv4Fallback(t *testing.T) {
	h := newTCPHarness(t, func(ctx context.Context, m *M.Metadata) (net.Conn, error) {
		if m.DstIP.Is6() {
			return nil, errors.New("exit has no IPv6")
		}
		return echoRemote(), nil
	})
	if c, err := h.dial("2001:db8::10"); err == nil {
		c.Close()
		t.Fatal("application falsely selected IPv6")
	}
	c, err := h.dial("192.0.2.10")
	if err != nil {
		t.Fatal(err)
	}
	assertEcho(t, c)
}

func TestStoppingStackOrTunnelCancelsPendingDial(t *testing.T) {
	for _, closeTunnel := range []bool{false, true} {
		t.Run(map[bool]string{false: "stack", true: "tunnel"}[closeTunnel], func(t *testing.T) {
			entered, exited := make(chan struct{}), make(chan struct{})
			h := newTCPHarness(t, func(ctx context.Context, _ *M.Metadata) (net.Conn, error) {
				close(entered)
				<-ctx.Done()
				close(exited)
				return nil, ctx.Err()
			})
			done := make(chan error, 1)
			go func() {
				c, err := h.dial("2001:db8::10")
				if c != nil {
					c.Close()
				}
				done <- err
			}()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("dial not entered")
			}
			if closeTunnel {
				h.deferred.Close()
			} else {
				h.cancel()
			}
			select {
			case <-exited:
			case <-time.After(time.Second):
				t.Fatal("dial not canceled")
			}
			if err := <-done; err == nil {
				t.Fatal("canceled connect succeeded")
			}
			if h.synACK.Load() != 0 {
				t.Fatal("SYNACK during stop")
			}
		})
	}
}

func TestConcurrentIPv4SuccessAndIPv6Failure(t *testing.T) {
	h := newTCPHarness(t, func(ctx context.Context, m *M.Metadata) (net.Conn, error) {
		if m.DstIP.Is6() {
			return nil, errors.New("IPv6 refused")
		}
		return echoRemote(), nil
	})
	var workers sync.WaitGroup
	for i := 0; i < 32; i++ {
		workers.Add(1)
		go func(i int) {
			defer workers.Done()
			ip := "192.0.2.10"
			if i%2 != 0 {
				ip = "2001:db8::10"
			}
			c, err := h.dial(ip)
			if i%2 != 0 {
				if c != nil {
					c.Close()
				}
				if err == nil {
					t.Error("IPv6 falsely connected")
				}
				return
			}
			if err != nil {
				t.Error(err)
				return
			}
			assertEcho(t, c)
		}(i)
	}
	workers.Wait()
}

func TestStoppingStackOrTunnelClosesEstablishedConnection(t *testing.T) {
	for _, closeTunnel := range []bool{false, true} {
		t.Run(map[bool]string{false: "stack", true: "tunnel"}[closeTunnel], func(t *testing.T) {
			remoteClosed := make(chan struct{})
			h := newTCPHarness(t, func(context.Context, *M.Metadata) (net.Conn, error) {
				a, b := net.Pipe()
				go func() {
					defer b.Close()
					defer close(remoteClosed)
					_, _ = io.Copy(io.Discard, b)
				}()
				return a, nil
			})
			c, err := h.dial("2001:db8::10")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			if closeTunnel {
				h.deferred.Close()
			} else {
				h.cancel()
			}
			select {
			case <-remoteClosed:
			case <-time.After(time.Second):
				t.Fatal("remote socket was not released on stop")
			}
			_ = c.SetReadDeadline(time.Now().Add(time.Second))
			_, err = c.Read(make([]byte, 1))
			if err == nil {
				t.Fatal("local socket remained open on stop")
			}
			if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
				t.Fatal("local socket timed out instead of closing")
			}
		})
	}
}
