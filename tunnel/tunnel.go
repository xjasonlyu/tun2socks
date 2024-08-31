package tunnel

import (
	"context"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/xjasonlyu/tun2socks/v2/core/adapter"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel/statistic"
)

const (
	// tcpConnectTimeout is the default timeout for TCP handshakes.
	tcpConnectTimeout = 5 * time.Second
	// tcpWaitTimeout implements a TCP half-close timeout.
	tcpWaitTimeout = 60 * time.Second
	// udpSessionTimeout is the default timeout for UDP sessions.
	udpSessionTimeout = 60 * time.Second
)

var _ adapter.TransportHandler = (*Tunnel)(nil)

type Tunnel struct {
	// Unbuffered TCP/UDP queues.
	tcpQueue chan adapter.TCPConn
	udpQueue chan adapter.UDPConn

	// UDP session timeout.
	udpTimeout *atomic.Duration

	// Internal proxy.Dialer for Tunnel.
	dialerMu sync.RWMutex
	dialer   proxy.Dialer

	// Where the Tunnel statistics are sent to.
	manager *statistic.Manager

	procOnce   sync.Once
	procCancel context.CancelFunc
}

func New(dialer proxy.Dialer, manager *statistic.Manager) *Tunnel {
	return &Tunnel{
		tcpQueue:   make(chan adapter.TCPConn),
		udpQueue:   make(chan adapter.UDPConn),
		udpTimeout: atomic.NewDuration(udpSessionTimeout),
		dialer:     dialer,
		manager:    manager,
		procCancel: func() { /* nop */ },
	}
}

// TCPIn return fan-in TCP queue.
func (t *Tunnel) TCPIn() chan<- adapter.TCPConn {
	return t.tcpQueue
}

// UDPIn return fan-in UDP queue.
func (t *Tunnel) UDPIn() chan<- adapter.UDPConn {
	return t.udpQueue
}

func (t *Tunnel) HandleTCP(conn adapter.TCPConn) {
	t.TCPIn() <- conn
}

func (t *Tunnel) HandleUDP(conn adapter.UDPConn) {
	t.UDPIn() <- conn
}

func (t *Tunnel) process(ctx context.Context) {
	for {
		select {
		case conn := <-t.tcpQueue:
			go t.handleTCPConn(conn)
		case conn := <-t.udpQueue:
			go t.handleUDPConn(conn)
		case <-ctx.Done():
			return
		}
	}
}

// ProcessAsync can be safely called multiple times, but will only be effective once.
func (t *Tunnel) ProcessAsync() {
	t.procOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		t.procCancel = cancel
		go t.process(ctx)
	})
}

// Close closes the Tunnel and releases its resources.
func (t *Tunnel) Close() {
	t.procCancel()
}

func (t *Tunnel) Dialer() proxy.Dialer {
	t.dialerMu.RLock()
	d := t.dialer
	t.dialerMu.RUnlock()
	return d
}

func (t *Tunnel) SetDialer(dialer proxy.Dialer) {
	t.dialerMu.Lock()
	t.dialer = dialer
	t.dialerMu.Unlock()
}

func (t *Tunnel) SetUDPTimeout(timeout time.Duration) {
	t.udpTimeout.Store(timeout)
}
