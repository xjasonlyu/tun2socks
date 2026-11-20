package socks5

import (
	"context"
	"io"
	"net"
	"net/netip"
	"testing"
	"time"

	M "github.com/xjasonlyu/tun2socks/v2/metadata"
)

func testSOCKSServer(t *testing.T, stage string, reply byte) (string, <-chan struct{}, <-chan struct{}) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	reached, closed := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(closed)
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_ = c.SetDeadline(time.Now().Add(3 * time.Second))
		greeting := make([]byte, 2)
		if _, err = io.ReadFull(c, greeting); err != nil {
			return
		}
		if _, err = io.ReadFull(c, make([]byte, int(greeting[1]))); err != nil {
			return
		}
		if stage == "greeting" {
			close(reached)
			_, _ = io.Copy(io.Discard, c)
			return
		}
		if stage == "authentication" {
			if _, err = c.Write([]byte{5, 2}); err != nil {
				return
			}
			auth := make([]byte, 2)
			if _, err = io.ReadFull(c, auth); err != nil {
				return
			}
			if _, err = io.ReadFull(c, make([]byte, int(auth[1]))); err != nil {
				return
			}
			if _, err = io.ReadFull(c, auth[:1]); err != nil {
				return
			}
			if _, err = io.ReadFull(c, make([]byte, int(auth[0]))); err != nil {
				return
			}
			close(reached)
			_, _ = io.Copy(io.Discard, c)
			return
		}
		if _, err = c.Write([]byte{5, 0}); err != nil {
			return
		}
		request := make([]byte, 4)
		if _, err = io.ReadFull(c, request); err != nil {
			return
		}
		size := 4 + 2
		if request[3] == 4 {
			size = 16 + 2
		}
		if _, err = io.ReadFull(c, make([]byte, size)); err != nil {
			return
		}
		close(reached)
		if stage == "connect" {
			_, _ = io.Copy(io.Discard, c)
			return
		}
		if _, err = c.Write([]byte{5, reply, 0, 1, 127, 0, 0, 1, 0, 80}); err != nil {
			return
		}
		if reply == 0 {
			_, _ = io.Copy(c, c)
		} else {
			_, _ = io.Copy(io.Discard, c)
		}
	}()
	return ln.Addr().String(), reached, closed
}

func TestLegacyDialContextRetainsHandshakeCancellationBehavior(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	reached, release, serverDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(serverDone)
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_ = c.SetDeadline(time.Now().Add(2 * time.Second))
		if _, err = io.ReadFull(c, make([]byte, 3)); err != nil {
			return
		}
		close(reached)
		<-release
	}()
	defer func() { close(release); <-serverDone }()
	p, err := New(ln.Addr().String(), "", "")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		c, err := p.DialContext(ctx, testMetadata())
		if c != nil {
			c.Close()
		}
		done <- err
	}()
	select {
	case <-reached:
	case <-time.After(time.Second):
		t.Fatal("handshake not reached")
	}
	cancel()
	select {
	case err := <-done:
		t.Fatalf("legacy handshake cancellation behavior changed: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	// Deferred cleanup closes the server, allowing the unchanged legacy read
	// to exit; the explicitly bounded method is tested separately above/below.
}

func testMetadata() *M.Metadata {
	return &M.Metadata{Network: M.TCP, DstIP: netip.MustParseAddr("2001:db8::10"), DstPort: 443}
}

func waitClosed(t *testing.T, closed <-chan struct{}) {
	t.Helper()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("proxy socket leaked")
	}
}

func TestConnectContextBoundsGreetingAndReply(t *testing.T) {
	for _, stage := range []string{"greeting", "connect"} {
		t.Run(stage, func(t *testing.T) {
			addr, _, closed := testSOCKSServer(t, stage, 0)
			p, _ := New(addr, "", "")
			ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
			defer cancel()
			started := time.Now()
			c, err := p.DialContextWithHandshake(ctx, testMetadata())
			if c != nil {
				c.Close()
			}
			if err == nil {
				t.Fatal("stalled handshake succeeded")
			}
			if time.Since(started) > time.Second {
				t.Fatal("connect deadline not enforced")
			}
			waitClosed(t, closed)
		})
	}
}

func TestConnectCancellationWithoutDeadlineClosesSocket(t *testing.T) {
	addr, reached, closed := testSOCKSServer(t, "connect", 0)
	p, _ := New(addr, "", "")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		c, err := p.DialContextWithHandshake(ctx, testMetadata())
		if c != nil {
			c.Close()
		}
		done <- err
	}()
	select {
	case <-reached:
	case <-time.After(time.Second):
		t.Fatal("CONNECT not received")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled dial succeeded")
		}
	case <-time.After(time.Second):
		t.Fatal("cancel did not interrupt handshake")
	}
	waitClosed(t, closed)
}

func TestSuccessfulConnectClearsDeadlineAndCancellationWatch(t *testing.T) {
	addr, _, closed := testSOCKSServer(t, "success", 0)
	p, _ := New(addr, "", "")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	c, err := p.DialContextWithHandshake(ctx, testMetadata())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	cancel() // The tunnel cancels its connect context as soon as DialTCP returns.
	time.Sleep(250 * time.Millisecond)
	if _, err := c.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	_ = c.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 5)
	if _, err := io.ReadFull(c, buf); err != nil || string(buf) != "hello" {
		t.Fatalf("echo=%q err=%v", buf, err)
	}
	c.Close()
	waitClosed(t, closed)
}

func TestSOCKSFailureReplyDoesNotReturnSuccess(t *testing.T) {
	addr, _, closed := testSOCKSServer(t, "reject", 8)
	p, _ := New(addr, "", "")
	c, err := p.DialContextWithHandshake(context.Background(), testMetadata())
	if c != nil {
		c.Close()
	}
	if err == nil {
		t.Fatal("SOCKS address-family rejection reported as success")
	}
	waitClosed(t, closed)
}
